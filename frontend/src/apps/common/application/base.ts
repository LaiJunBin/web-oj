import { useLoadingStorage } from '@apps/common/domain/loading';

export class Application {
	withLoading() {
		const loadingStorage = useLoadingStorage();
		return new Proxy(this, {
			get: (target, prop) => {
				const origProp = target[prop as keyof Application];
				if (typeof origProp == 'function') {
					loadingStorage.increase();
					return function (...args: any) {
						try {
							let result = origProp.apply(target, args);
							return Promise.resolve(result).finally(() => {
								loadingStorage.decrease();
							});
						} catch (error) {
							console.error(error);
							throw error;
						}
					};
				}

				return origProp;
			}
		});
	}
}
