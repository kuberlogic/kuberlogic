/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { OnDestroy } from '@angular/core';
import { Constructor } from '@app/mixins/mixins';
import { Observable, Subject } from 'rxjs';

// eslint-disable-next-line
export function DestroyMixin<TBase extends Constructor>(Base: TBase) {
    abstract class DestroyMixinClass extends Base implements OnDestroy {
        protected destroyed$: Observable<void>;
        private destroySource = new Subject<void>();

        protected constructor(...args: any[]) {
            super(args);
            this.destroyed$ = this.destroySource.asObservable();
        }

        ngOnDestroy(): void {
            this.destroySource.next();
            this.destroySource.complete();
        }
    }

    return DestroyMixinClass;
}
