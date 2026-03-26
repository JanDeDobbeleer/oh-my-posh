import { query } from '../db';

export interface UserRow {
  id: string; email: string; password_hash: string;
  first_name: string; last_name: string; phone: string | null;
  status: 'active' | 'inactive' | 'suspended' | 'pending_verification';
  role: 'user' | 'admin'; created_at: Date;
}
export type UserPublic = Omit<UserRow, 'password_hash'>;
export interface CreateUserInput { email: string; password_hash: string; first_name: string; last_name: string; phone?: string; role?: 'user' | 'admin'; }
export interface UpdateUserInput { first_name?: string; last_name?: string; phone?: string; }

export async function findById(id: string): Promise<UserRow | null> {
  const rows = await query<UserRow>('SELECT id,email,password_hash,first_name,last_name,phone,status,role,created_at FROM users WHERE id=$1', [id]);
  return rows[0] ?? null;
}
export async function findByEmail(email: string): Promise<UserRow | null> {
  const rows = await query<UserRow>('SELECT id,email,password_hash,first_name,last_name,phone,status,role,created_at FROM users WHERE email=$1', [email]);
  return rows[0] ?? null;
}
export async function create(input: CreateUserInput): Promise<UserRow> {
  const rows = await query<UserRow>(
    'INSERT INTO users (email,password_hash,first_name,last_name,phone,role) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id,email,password_hash,first_name,last_name,phone,status,role,created_at',
    [input.email, input.password_hash, input.first_name, input.last_name, input.phone ?? null, input.role ?? 'user']
  );
  if (!rows[0]) throw new Error('Failed to create user');
  return rows[0];
}
export async function update(id: string, input: UpdateUserInput): Promise<UserRow | null> {
  const fields: string[] = []; const values: unknown[] = []; let i = 1;
  if (input.first_name !== undefined) { fields.push(`first_name=$${i++}`); values.push(input.first_name); }
  if (input.last_name !== undefined)  { fields.push(`last_name=$${i++}`);  values.push(input.last_name); }
  if (input.phone !== undefined)      { fields.push(`phone=$${i++}`);      values.push(input.phone); }
  if (!fields.length) return findById(id);
  values.push(id);
  const rows = await query<UserRow>(`UPDATE users SET ${fields.join(',')},updated_at=NOW() WHERE id=$${i} RETURNING id,email,password_hash,first_name,last_name,phone,status,role,created_at`, values);
  return rows[0] ?? null;
}
export async function updateStatus(id: string, status: UserRow['status']): Promise<UserRow | null> {
  const rows = await query<UserRow>('UPDATE users SET status=$1,updated_at=NOW() WHERE id=$2 RETURNING id,email,password_hash,first_name,last_name,phone,status,role,created_at', [status, id]);
  return rows[0] ?? null;
}
export function toPublic(user: UserRow): UserPublic {
  const { password_hash, ...pub } = user; void password_hash; return pub;
}
