import { initAdminContainer } from './admin/container';
import { initCommonContainer } from './common/container';
import { initContestContainer } from './contest/container';

export function initApps() {
	initCommonContainer();
	initContestContainer();
	initAdminContainer();
}
