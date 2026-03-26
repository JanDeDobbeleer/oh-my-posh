import { Router, Request, Response, NextFunction } from 'express';
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { z } from 'zod';
import { validate } from '../middleware/validate';
import * as UserModel from '../models/User';
import * as AccountModel from '../models/Account';
import { UnauthorizedError, ConflictError } from '../middleware/errorHandler';

const router = Router();
const refreshTokenStore = new Set<string>();

const loginSchema = z.object({ email: z.string().email(), password: z.string().min(1) });
const registerSchema = z.object({
  email: z.string().email(), password: z.string().min(8).regex(/[A-Z]/).regex(/[0-9]/),
  first_name: z.string().min(1).max(100), last_name: z.string().min(1).max(100), phone: z.string().optional()
});
const refreshSchema = z.object({ refresh_token: z.string().min(1) });

function generateTokens(user: UserModel.UserRow) {
  const payload = { id: user.id, email: user.email, role: user.role };
  return {
    accessToken:  jwt.sign(payload, process.env.JWT_SECRET!,         { expiresIn: '24h' }),
    refreshToken: jwt.sign(payload, process.env.JWT_REFRESH_SECRET!, { expiresIn: '7d'  }),
  };
}

router.post('/login', validate(loginSchema), async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { email, password } = req.body as z.infer<typeof loginSchema>;
    const user = await UserModel.findByEmail(email);
    if (!user || user.status !== 'active') throw new UnauthorizedError('Invalid credentials');
    if (!await bcrypt.compare(password, user.password_hash)) throw new UnauthorizedError('Invalid credentials');
    const { accessToken, refreshToken } = generateTokens(user);
    refreshTokenStore.add(refreshToken);
    res.json({ success: true, data: { access_token: accessToken, refresh_token: refreshToken, token_type: 'Bearer', user: UserModel.toPublic(user) } });
  } catch (e) { next(e); }
});

router.post('/register', validate(registerSchema), async (req: Request, res: Response, next: NextFunction) => {
  try {
    const input = req.body as z.infer<typeof registerSchema>;
    if (await UserModel.findByEmail(input.email)) throw new ConflictError('Email already exists');
    const password_hash = await bcrypt.hash(input.password, 12);
    const user = await UserModel.create({ email: input.email, password_hash, first_name: input.first_name, last_name: input.last_name, phone: input.phone });
    await AccountModel.create({ user_id: user.id });
    const { accessToken, refreshToken } = generateTokens(user);
    refreshTokenStore.add(refreshToken);
    res.status(201).json({ success: true, data: { access_token: accessToken, refresh_token: refreshToken, token_type: 'Bearer', user: UserModel.toPublic(user) } });
  } catch (e) { next(e); }
});

router.post('/refresh', validate(refreshSchema), async (req: Request, res: Response, next: NextFunction) => {
  try {
    const { refresh_token } = req.body as z.infer<typeof refreshSchema>;
    if (!refreshTokenStore.has(refresh_token)) throw new UnauthorizedError('Invalid refresh token');
    const decoded = jwt.verify(refresh_token, process.env.JWT_REFRESH_SECRET!) as UserModel.UserRow;
    const user = await UserModel.findById(decoded.id);
    if (!user || user.status !== 'active') throw new UnauthorizedError('User not found');
    refreshTokenStore.delete(refresh_token);
    const { accessToken, refreshToken: newRefresh } = generateTokens(user);
    refreshTokenStore.add(newRefresh);
    res.json({ success: true, data: { access_token: accessToken, refresh_token: newRefresh, token_type: 'Bearer' } });
  } catch (e) { next(e); }
});

router.post('/logout', validate(refreshSchema), (req: Request, res: Response) => {
  refreshTokenStore.delete(req.body.refresh_token);
  res.json({ success: true, data: { message: 'Logged out' } });
});

export default router;
