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
import { ServiceDatabaseModel } from '@models/service-database.model';
import { MessagesService } from '@services/messages.service';
import { BehaviorSubject, Observable } from 'rxjs';
import { catchError, mergeMap, publishLast, refCount, tap } from 'rxjs/operators';

@Injectable({
    providedIn: 'root'
})
export class ServiceDatabasesService {
    private readonly databases$: Observable<ServiceDatabaseModel[] | undefined>;
    private databasesSource = new BehaviorSubject<ServiceDatabaseModel[] | undefined>(undefined);
    private databasesCache$: Observable<ServiceDatabaseModel[] | undefined> | undefined;
    private currentServiceId = '';

    constructor(
        private http: HttpClient,
        private messages: MessagesService,
    ) {
        this.databases$ = this.databasesSource.asObservable();
    }

    getDatabases(serviceId: string): Observable<ServiceDatabaseModel[] | undefined> {
        return this.loadDatabases(serviceId).pipe(
            mergeMap(() => {
                return this.databases$;
            }),
        );
    }

    createDatabase(serviceId: string, db: ServiceDatabaseModel): Observable<ServiceDatabaseModel> {
        return this.http
            .post<ServiceDatabaseModel>(`${environment.apiUrl}/services/${serviceId}/databases`, db)
            .pipe(
                tap(() => {
                    this.reloadDatabases();
                }),
            );
    }

    deleteDatabase(serviceId: string, name: string): Observable<ServiceDatabaseModel> {
        return this.http
            .delete<ServiceDatabaseModel>(`${environment.apiUrl}/services/${serviceId}/databases/${name}`)
            .pipe(
                tap(() => {
                    this.reloadDatabases();
                }),
            );
    }

    private reloadDatabases(): void {
        this.databasesCache$ = undefined;
        if (!!this.currentServiceId) {
            this.loadDatabases(this.currentServiceId).subscribe();
        }
    }

    private loadDatabases(serviceId: string): Observable<ServiceDatabaseModel[] | undefined> {
        if (serviceId !== this.currentServiceId) {
            this.databasesCache$ = undefined;
            this.currentServiceId = serviceId;
        }

        if (!this.databasesCache$) {
            this.databasesCache$ = this.http
                .get<ServiceDatabaseModel[] | undefined>(`${environment.apiUrl}/services/${serviceId}/databases`)
                .pipe(
                    publishLast(),
                    refCount(),
                    catchError((err) => {
                        this.messages.error(err);
                        this.databasesCache$ = undefined;
                        throw err;
                    }),
                );
        }
        return this.databasesCache$.pipe(
            tap((users) => {
                this.databasesSource.next(users);
            })
        );
    }
}
