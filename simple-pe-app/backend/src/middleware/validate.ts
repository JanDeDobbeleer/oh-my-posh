import { Request, Response, NextFunction } from 'express';
import { ZodSchema, ZodError } from 'zod';

export function validate(schema: ZodSchema) {
  return (req: Request, res: Response, next: NextFunction): void => {
    const result = schema.safeParse(req.body);
    if (!result.success) {
      res.status(400).json({ success: false, error: { code: 'VALIDATION_ERROR', message: 'Validation failed', details: result.error.issues } });
      return;
    }
    req.body = result.data; next();
  };
}

export function validateQuery(schema: ZodSchema) {
  return (req: Request, res: Response, next: NextFunction): void => {
    const result = schema.safeParse(req.query);
    if (!result.success) {
      res.status(400).json({ success: false, error: { code: 'VALIDATION_ERROR', message: 'Query validation failed', details: result.error.issues } });
      return;
    }
    req.query = result.data as Record<string, string>; next();
  };
}
