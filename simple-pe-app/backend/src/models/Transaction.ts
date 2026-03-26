import { query, getClient } from '../db';

export interface TransactionRow {
  id: string; from_account_id: string | null; to_account_id: string | null;
  amount: string; currency: string;
  type: 'transfer' | 'deposit' | 'withdrawal' | 'payment' | 'refund';
  status: 'pending' | 'completed' | 'failed' | 'cancelled' | 'reversed';
  description: string | null; metadata: Record<string, unknown> | null; created_at: Date;
}
export interface CreateTransactionInput {
  from_account_id?: string; to_account_id?: string; amount: number;
  currency?: string; type: TransactionRow['type']; description?: string;
  metadata?: Record<string, unknown>;
}

export async function findByAccountId(filters: { accountId: string; startDate?: string; endDate?: string; type?: string; status?: string; limit?: number; offset?: number }): Promise<TransactionRow[]> {
  const conds = ['(from_account_id=$1 OR to_account_id=$1)']; const vals: unknown[] = [filters.accountId]; let i = 2;
  if (filters.startDate) { conds.push(`created_at>=$${i++}`); vals.push(filters.startDate); }
  if (filters.endDate)   { conds.push(`created_at<=$${i++}`); vals.push(filters.endDate); }
  if (filters.type)      { conds.push(`type=$${i++}`); vals.push(filters.type); }
  if (filters.status)    { conds.push(`status=$${i++}`); vals.push(filters.status); }
  vals.push(filters.limit ?? 20, filters.offset ?? 0);
  return query<TransactionRow>(`SELECT id,from_account_id,to_account_id,amount,currency,type,status,description,metadata,created_at FROM transactions WHERE ${conds.join(' AND ')} ORDER BY created_at DESC LIMIT $${i++} OFFSET $${i}`, vals);
}
export async function findById(id: string): Promise<TransactionRow | null> {
  const rows = await query<TransactionRow>('SELECT id,from_account_id,to_account_id,amount,currency,type,status,description,metadata,created_at FROM transactions WHERE id=$1', [id]);
  return rows[0] ?? null;
}
export async function create(input: CreateTransactionInput): Promise<TransactionRow> {
  const rows = await query<TransactionRow>(
    'INSERT INTO transactions (from_account_id,to_account_id,amount,currency,type,description,metadata) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id,from_account_id,to_account_id,amount,currency,type,status,description,metadata,created_at',
    [input.from_account_id ?? null, input.to_account_id ?? null, input.amount, input.currency ?? 'PEN', input.type, input.description ?? null, input.metadata ? JSON.stringify(input.metadata) : null]
  );
  if (!rows[0]) throw new Error('Failed to create transaction');
  return rows[0];
}
export async function createWithClient(client: Awaited<ReturnType<typeof getClient>>, input: CreateTransactionInput): Promise<TransactionRow> {
  const res = await client.query<TransactionRow>(
    'INSERT INTO transactions (from_account_id,to_account_id,amount,currency,type,description,metadata) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id,from_account_id,to_account_id,amount,currency,type,status,description,metadata,created_at',
    [input.from_account_id ?? null, input.to_account_id ?? null, input.amount, input.currency ?? 'PEN', input.type, input.description ?? null, input.metadata ? JSON.stringify(input.metadata) : null]
  );
  if (!res.rows[0]) throw new Error('Failed to create transaction');
  return res.rows[0];
}
export async function getMonthlySummary(accountId: string, year: number, month: number): Promise<{ income: string; expenses: string; count: string }> {
  const start = `${year}-${String(month).padStart(2,'0')}-01`;
  const end   = `${year}-${String(month+1).padStart(2,'0')}-01`;
  const rows = await query<{ income: string; expenses: string; count: string }>(
    `SELECT COALESCE(SUM(CASE WHEN to_account_id=$1 AND status='completed' THEN amount ELSE 0 END),0) AS income, COALESCE(SUM(CASE WHEN from_account_id=$1 AND status='completed' THEN amount ELSE 0 END),0) AS expenses, COUNT(*) AS count FROM transactions WHERE (from_account_id=$1 OR to_account_id=$1) AND created_at>=$2 AND created_at<$3`,
    [accountId, start, end]
  );
  return rows[0] ?? { income: '0', expenses: '0', count: '0' };
}
