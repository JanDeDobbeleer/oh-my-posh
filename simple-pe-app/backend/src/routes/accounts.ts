import { Router, Request, Response, NextFunction } from 'express';
import { authenticate } from '../middleware/auth';
import * as AccountModel from '../models/Account';
import { NotFoundError, UnauthorizedError } from '../middleware/errorHandler';

const router = Router();
router.use(authenticate);

router.get('/', async (req: Request, res: Response, next: NextFunction) => {
  try { res.json({ success: true, data: await AccountModel.findByUserId(req.user!.id) }); } catch (e) { next(e); }
});

router.get('/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const account = await AccountModel.findById(req.params.id);
    if (!account) throw new NotFoundError('Account not found');
    if (account.user_id !== req.user!.id && req.user!.role !== 'admin') throw new UnauthorizedError('Access denied');
    res.json({ success: true, data: account });
  } catch (e) { next(e); }
});

router.get('/:id/balance', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const account = await AccountModel.findById(req.params.id);
    if (!account) throw new NotFoundError('Account not found');
    if (account.user_id !== req.user!.id && req.user!.role !== 'admin') throw new UnauthorizedError('Access denied');
    res.json({ success: true, data: { account_id: account.id, balance: account.balance, currency: account.currency } });
  } catch (e) { next(e); }
});

export default router;
