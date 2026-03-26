import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';

export interface JwtPayload { id: string; email: string; role: string; }

declare global { namespace Express { interface Request { user?: JwtPayload; } } }

export function authenticate(req: Request, res: Response, next: NextFunction): void {
  const authHeader = req.headers.authorization;
  if (!authHeader?.startsWith('Bearer ')) {
    res.status(401).json({ success: false, error: { code: 'TOKEN_MISSING', message: 'Authorization token is required' } });
    return;
  }
  const token = authHeader.split(' ')[1];
  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET!) as JwtPayload;
    req.user = decoded;
    next();
  } catch (error) {
    const code = error instanceof jwt.TokenExpiredError ? 'TOKEN_EXPIRED' : 'TOKEN_INVALID';
    res.status(403).json({ success: false, error: { code, message: 'Invalid or expired token' } });
  }
}

export function requireRole(...roles: string[]) {
  return (req: Request, res: Response, next: NextFunction): void => {
    if (!req.user) { res.status(401).json({ success: false, error: { code: 'UNAUTHORIZED' } }); return; }
    if (!roles.includes(req.user.role)) { res.status(403).json({ success: false, error: { code: 'FORBIDDEN' } }); return; }
    next();
  };
}
