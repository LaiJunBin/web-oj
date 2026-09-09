import type { z } from 'zod';
import { UserProjectApiSchemas } from '@/api/gen/UserProject';
import type { Storage } from '@apps/common/domain/storage';
import { container } from '@/apps/contest/container';

export type Scoreboard = z.infer<typeof UserProjectApiSchemas.User_ProjectSpec_ProjectScoreboard>;

export interface ScoreboardStorageInterface extends Storage<Scoreboard> {}
export const useScoreboardStorage = () =>
	container.get('ScoreboardStorage') as ScoreboardStorageInterface;
