import { Router, Request, Response, NextFunction } from 'express';
import { z } from 'zod';
import { authenticate, requireRole } from '../middleware/auth';
import { validate } from '../middleware/validate';
import * as UserModel from '../models/User';
import { NotFoundError } from '../middleware/errorHandler';

const router = Router();
router.use(authenticate);

const updateMeSchema = z.object({ first_name: z.string().min(1).max(100).optional(), last_name: z.string().min(1).max(100).optional(), phone: z.string().optional() });
const updateStatusSchema = z.object({ status: z.enum(['active', 'inactive', 'suspended', 'pending_verification']) });

router.get('/me', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const user = await UserModel.findById(req.user!.id);
    if (!user) throw new NotFoundError('User not found');
    res.json({ success: true, data: UserModel.toPublic(user) });
  } catch (e) { next(e); }
});

router.put('/me', validate(updateMeSchema), async (req: Request, res: Response, next: NextFunction) => {
  try {
    const updated = await UserModel.update(req.user!.id, req.body);
    if (!updated) throw new NotFoundError('User not found');
    res.json({ success: true, data: UserModel.toPublic(updated) });
  } catch (e) { next(e); }
});

router.get('/:id', requireRole('admin'), async (req: Request, res: Response, next: NextFunction) => {
  try {
    const user = await UserModel.findById(req.params.id);
    if (!user) throw new NotFoundError('User not found');
    res.json({ success: true, data: UserModel.toPublic(user) });
  } catch (e) { next(e); }
});

router.patch('/:id/status', requireRole('admin'), validate(updateStatusSchema), async (req: Request, res: Response, next: NextFunction) => {
  try {
    const updated = await UserModel.updateStatus(req.params.id, req.body.status);
    if (!updated) throw new NotFoundError('User not found');
    res.json({ success: true, data: UserModel.toPublic(updated) });
  } catch (e) { next(e); }
});

export default router;
