import { z, ZodObject } from 'zod';
import { BuildsApiSchemas } from '@/api/gen/Builds';

export type BuildInfo = z.infer<typeof BuildsApiSchemas.Admin_BuildSpec_BuildInfo>;
export type Build = z.infer<typeof BuildsApiSchemas.Admin_BuildSpec_Build>;
export type BuildLog = z.infer<typeof BuildsApiSchemas.Admin_BuildSpec_Log>;

export const BuildLogEventDataSchema = (
	BuildsApiSchemas.Admin_BuildSpec_Log as ZodObject<any, any, any, any, any>
).pick({
	level: true,
	message: true,
	time: true
}) as ZodObject<any, 'strip', any, BuildLog, any>;
