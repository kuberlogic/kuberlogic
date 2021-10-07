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
import { ServiceRestoreModel } from '@models/service-restore.model';
import { MessagesService } from '@services/messages.service';
import { BehaviorSubject, Observable } from 'rxjs';
import { catchError, mergeMap, publishLast, refCount, tap } from 'rxjs/operators';

@Injectable({
    providedIn: 'root'
})
export class ServiceRestoresService {

    private readonly data$: Observable<ServiceRestoreModel[] | undefined>;
    private dataSource = new BehaviorSubject<ServiceRestoreModel[] | undefined>(undefined);
    private cache$: Observable<ServiceRestoreModel[]> | undefined;
    private currentServiceId = '';

    constructor(
        private http: HttpClient,
        private messages: MessagesService,
    ) {
        this.data$ = this.dataSource.asObservable();
    }

    getList(serviceId: string): Observable<ServiceRestoreModel[] | undefined> {
        return this.load(serviceId).pipe(
            mergeMap(() => this.data$),
        );
    }

    restore(serviceId: string, key: string, database: string): Observable<string> {
        return this.http
            .post<string>(`${environment.apiUrl}/services/${serviceId}/restores`, {key, database})
            .pipe(
                tap(() => {
                    this.reload();
                }),
            );
    }

    private load(serviceId: string): Observable<ServiceRestoreModel[]> {
        if (serviceId !== this.currentServiceId) {
            this.cache$ = undefined;
            this.currentServiceId = serviceId;
        }

        if (!this.cache$) {
            this.cache$ = this.http.get<ServiceRestoreModel[]>(`${environment.apiUrl}/services/${serviceId}/restores`)
                .pipe(
                    publishLast(),
                    refCount(),
                    catchError((err) => {
                        this.messages.error(err);
                        this.cache$ = undefined;
                        throw err;
                    }),
                );
        }
        return this.cache$.pipe(
            tap((models) => {
                this.dataSource.next(models);
                return models;
            })
        );
    }

    private reload(): void {
        this.cache$ = undefined;
        if (!!this.currentServiceId) {
            this.load(this.currentServiceId).subscribe();
        }
    }
}
