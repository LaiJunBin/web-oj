import { setupServer } from 'msw/node'
import { handlers } from './handlers.overwritten'

export const server = setupServer(...handlers)
