import type { RolesApiSchemas } from '@/api/gen/Roles';
import type { z } from 'zod';
import type { Storage } from '@/apps/common/domain/storage';
import { container } from '@/apps/admin/container';

export type Role = z.infer<typeof RolesApiSchemas.Admin_RoleSpec_Role>;

export interface RolesStorageInterface extends Storage<Role[]> {}
export const useRolesStorage = () => container.get('RolesStorage') as RolesStorageInterface;
