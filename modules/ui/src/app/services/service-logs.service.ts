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

import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from '@environments/environment';
import { ServiceLogModel } from '@models/service-log.model';
import { BehaviorSubject, Observable } from 'rxjs';
import { mergeMap, publishLast, refCount, tap } from 'rxjs/operators';

@Injectable({
    providedIn: 'root'
})
export class ServiceLogsService {

    private readonly data$: Observable<ServiceLogModel | undefined>;
    private dataSource = new BehaviorSubject<ServiceLogModel | undefined>(undefined);

    constructor(
        private http: HttpClient,
    ) {
        this.data$ = this.dataSource.asObservable();
    }

    get(serviceId: string, serviceInstance: string): Observable<ServiceLogModel | undefined> {
        return this.load(serviceId, serviceInstance).pipe(
            mergeMap(() => {
                return this.data$;
            }),
        );
    }

    private load(serviceId: string, serviceInstance: string): Observable<ServiceLogModel | undefined> {
        return this.http
            .get<ServiceLogModel | undefined>(`${environment.apiUrl}/services/${serviceId}/logs`,
                {params: {service_instance: serviceInstance}})
            .pipe(
                publishLast(),
                refCount(),
                tap((data) => this.dataSource.next(data)),
            );
    }
}
