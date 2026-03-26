import { query, getClient } from '../db';

export interface AccountRow {
  id: string; user_id: string; account_number: string; balance: string;
  currency: string; type: 'savings' | 'current' | 'investment';
  status: 'active' | 'inactive' | 'frozen'; created_at: Date;
}
export interface CreateAccountInput { user_id: string; currency?: string; type?: 'savings' | 'current' | 'investment'; }

export async function findByUserId(userId: string): Promise<AccountRow[]> {
  return query<AccountRow>('SELECT id,user_id,account_number,balance,currency,type,status,created_at FROM accounts WHERE user_id=$1 ORDER BY created_at ASC', [userId]);
}
export async function findById(id: string): Promise<AccountRow | null> {
  const rows = await query<AccountRow>('SELECT id,user_id,account_number,balance,currency,type,status,created_at FROM accounts WHERE id=$1', [id]);
  return rows[0] ?? null;
}
export async function create(input: CreateAccountInput): Promise<AccountRow> {
  const rows = await query<AccountRow>(
    'INSERT INTO accounts (user_id,currency,type) VALUES ($1,$2,$3) RETURNING id,user_id,account_number,balance,currency,type,status,created_at',
    [input.user_id, input.currency ?? 'PEN', input.type ?? 'savings']
  );
  if (!rows[0]) throw new Error('Failed to create account');
  return rows[0];
}
export async function updateBalance(fromId: string, toId: string, amount: number): Promise<{ from: AccountRow; to: AccountRow }> {
  const client = await getClient();
  try {
    await client.query('BEGIN');
    const fromR = await client.query<AccountRow>('SELECT * FROM accounts WHERE id=$1 FOR UPDATE', [fromId]);
    const toR   = await client.query<AccountRow>('SELECT * FROM accounts WHERE id=$1 FOR UPDATE', [toId]);
    if (!fromR.rows[0]) throw new Error('Source account not found');
    if (!toR.rows[0])   throw new Error('Dest account not found');
    if (parseFloat(fromR.rows[0].balance) < amount) throw new Error('Insufficient funds');
    const upFrom = await client.query<AccountRow>('UPDATE accounts SET balance=balance-$1,updated_at=NOW() WHERE id=$2 RETURNING *', [amount, fromId]);
    const upTo   = await client.query<AccountRow>('UPDATE accounts SET balance=balance+$1,updated_at=NOW() WHERE id=$2 RETURNING *', [amount, toId]);
    await client.query('COMMIT');
    return { from: upFrom.rows[0]!, to: upTo.rows[0]! };
  } catch (e) { await client.query('ROLLBACK'); throw e; }
  finally { client.release(); }
}
