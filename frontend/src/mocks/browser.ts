import { setupWorker } from 'msw/browser';
import { handlers } from './handlers.overwritten';

export const worker = setupWorker(...handlers);
