import express, { Request, Response } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import morgan from 'morgan';
import dotenv from 'dotenv';
dotenv.config();

import authRouter from './routes/auth';
import usersRouter from './routes/users';
import accountsRouter from './routes/accounts';
import transactionsRouter from './routes/transactions';
import { errorHandler } from './middleware/errorHandler';
import { pool } from './db';

const app = express();
app.use(helmet());
app.use(cors({ origin: process.env.FRONTEND_URL ?? 'http://localhost:5173', credentials: true }));
app.use(morgan(process.env.NODE_ENV === 'production' ? 'combined' : 'dev'));
app.use(express.json({ limit: '10mb' }));

app.get('/health', async (_req: Request, res: Response) => {
  try { await pool.query('SELECT 1'); res.json({ status: 'ok', database: 'connected' }); }
  catch { res.status(503).json({ status: 'error', database: 'disconnected' }); }
});

app.use('/api/v1/auth', authRouter);
app.use('/api/v1/users', usersRouter);
app.use('/api/v1/accounts', accountsRouter);
app.use('/api/v1/transactions', transactionsRouter);

app.use((_req, res) => res.status(404).json({ success: false, error: { code: 'NOT_FOUND', message: 'Endpoint not found' } }));
app.use(errorHandler);

const PORT = parseInt(process.env.PORT ?? '4000', 10);
app.listen(PORT, () => console.log(`[server] http://localhost:${PORT}/api/v1`));
export default app;
