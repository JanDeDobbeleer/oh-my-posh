import { Request, Response, NextFunction } from 'express';

export class AppError extends Error {
  constructor(message: string, public readonly statusCode: number, public readonly code: string) {
    super(message); this.name = this.constructor.name;
  }
}
export class ValidationError extends AppError { constructor(msg: string) { super(msg, 400, 'VALIDATION_ERROR'); } }
export class UnauthorizedError extends AppError { constructor(msg = 'Unauthorized') { super(msg, 401, 'UNAUTHORIZED'); } }
export class NotFoundError extends AppError { constructor(msg = 'Not found') { super(msg, 404, 'NOT_FOUND'); } }
export class ConflictError extends AppError { constructor(msg: string) { super(msg, 409, 'CONFLICT'); } }
export class InsufficientFundsError extends AppError { constructor(msg = 'Insufficient funds') { super(msg, 422, 'INSUFFICIENT_FUNDS'); } }

export function errorHandler(err: Error, req: Request, res: Response, _next: NextFunction): void {
  console.error(err.message, { path: req.path });
  if (err instanceof AppError) {
    res.status(err.statusCode).json({ success: false, error: { code: err.code, message: err.message } }); return;
  }
  res.status(500).json({ success: false, error: { code: 'INTERNAL_SERVER_ERROR', message: process.env.NODE_ENV === 'production' ? 'Unexpected error' : err.message } });
}
