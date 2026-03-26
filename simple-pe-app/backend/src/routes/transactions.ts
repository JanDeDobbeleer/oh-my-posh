import { Router, Request, Response, NextFunction } from 'express';
import { z } from 'zod';
import { authenticate } from '../middleware/auth';
import { validate, validateQuery } from '../middleware/validate';
import * as TransactionModel from '../models/Transaction';
import * as AccountModel from '../models/Account';
import { getClient } from '../db';
import { NotFoundError, UnauthorizedError, ValidationError, InsufficientFundsError } from '../middleware/errorHandler';

const router = Router();
router.use(authenticate);

const listQuerySchema = z.object({
  page: z.string().regex(/^\d+$/).transform(Number).optional(),
  limit: z.string().regex(/^\d+$/).transform(Number).optional(),
  type: z.enum(['transfer','deposit','withdrawal','payment','refund']).optional(),
  status: z.enum(['pending','completed','failed','cancelled','reversed']).optional(),
  account_id: z.string().uuid().optional(),
});

const transferSchema = z.object({
  from_account_id: z.string().uuid(),
  to_account_id: z.string().uuid(),
  amount: z.number().positive().max(50000),
  currency: z.string().length(3).optional().default('PEN'),
  description: z.string().max(255).optional(),
});

router.get('/', validateQuery(listQuerySchema), async (req: Request, res: Response, next: NextFunction) => {
  try {
    const p = req.query as z.infer<typeof listQuerySchema>;
    const page = Number(p.page ?? 1); const limit = Math.min(Number(p.limit ?? 20), 100); const offset = (page-1)*limit;
    const accounts = await AccountModel.findByUserId(req.user!.id);
    if (!accounts.length) return res.json({ success: true, data: [] });
    const accountId = p.account_id ?? accounts[0]!.id;
    const txs = await TransactionModel.findByAccountId({ accountId, type: p.type, status: p.status, limit, offset });
    return res.json({ success: true, data: txs, pagination: { page, limit, has_more: txs.length === limit } });
  } catch (e) { return next(e); }
});

router.get('/summary', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const now = new Date(); const accounts = await AccountModel.findByUserId(req.user!.id);
    if (!accounts.length) return res.json({ success: true, data: { income: '0', expenses: '0', count: '0' } });
    const summary = await TransactionModel.getMonthlySummary(accounts[0]!.id, now.getFullYear(), now.getMonth()+1);
    return res.json({ success: true, data: summary });
  } catch (e) { return next(e); }
});

router.get('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const tx = await TransactionModel.findById(req.params.id);
    if (!tx) throw new NotFoundError('Transaction not found');
    const accounts = await AccountModel.findByUserId(req.user!.id);
    const ids = new Set(accounts.map(a => a.id));
    if (!((tx.from_account_id && ids.has(tx.from_account_id)) || (tx.to_account_id && ids.has(tx.to_account_id)) || req.user!.role === 'admin')) throw new UnauthorizedError('Access denied');
    res.json({ success: true, data: tx });
  } catch (e) { next(e); }
});

router.post('/transfer', validate(transferSchema), async (req: Request, res: Response, next: NextFunction) => {
  const client = await getClient();
  try {
    const input = req.body as z.infer<typeof transferSchema>;
    if (input.from_account_id === input.to_account_id) throw new ValidationError('Same account');
    const fromAccount = await AccountModel.findById(input.from_account_id);
    if (!fromAccount) throw new NotFoundError('Source account not found');
    if (fromAccount.user_id !== req.user!.id) throw new UnauthorizedError('Not your account');
    if (fromAccount.status !== 'active') throw new ValidationError('Source account inactive');
    const toAccount = await AccountModel.findById(input.to_account_id);
    if (!toAccount) throw new NotFoundError('Destination account not found');
    if (parseFloat(fromAccount.balance) < input.amount) throw new InsufficientFundsError();
    await client.query('BEGIN');
    const lock = await client.query<{ balance: string }>('SELECT balance FROM accounts WHERE id=$1 FOR UPDATE', [input.from_account_id]);
    if (parseFloat(lock.rows[0]?.balance ?? '0') < input.amount) throw new InsufficientFundsError();
    await client.query('UPDATE accounts SET balance=balance-$1,updated_at=NOW() WHERE id=$2', [input.amount, input.from_account_id]);
    await client.query('UPDATE accounts SET balance=balance+$1,updated_at=NOW() WHERE id=$2', [input.amount, input.to_account_id]);
    const tx = await TransactionModel.createWithClient(client, { from_account_id: input.from_account_id, to_account_id: input.to_account_id, amount: input.amount, currency: input.currency, type: 'transfer', description: input.description });
    await client.query("UPDATE transactions SET status='completed' WHERE id=$1", [tx.id]);
    await client.query('COMMIT');
    tx.status = 'completed';
    res.status(201).json({ success: true, data: tx });
  } catch (e) { await client.query('ROLLBACK'); next(e); }
  finally { client.release(); }
});

export default router;
