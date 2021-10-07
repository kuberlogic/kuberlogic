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
import { Validators } from '@angular/forms';
import { environment } from '@environments/environment';
import { ServiceModel } from '@models/service.model';
import { MessagesService } from '@services/messages.service';
import { BehaviorSubject, Observable } from 'rxjs';
import { catchError, filter, map, mergeMap, publishLast, refCount, tap } from 'rxjs/operators';

export type ServiceInstancesNames = ReadonlyMap<string, string>;

const floatingNumberPattern = '^[0-9]+\\.?[0-9]*$';
export const limitFormGroup = {
    cpu: ['', [
        Validators.required,
        Validators.min(0.25),
        Validators.pattern(floatingNumberPattern),
    ]],
    memory: ['', [
        Validators.required,
        Validators.min(0.5),
        Validators.pattern(floatingNumberPattern),
    ]],
    volumeSize: ['', [
        Validators.required,
        Validators.min(1),
        Validators.pattern(floatingNumberPattern),
    ]],
};

@Injectable({
    providedIn: 'root'
})
export class ServicesPageService {
    private readonly services$: Observable<ServiceModel[] | undefined>;
    private servicesSource = new BehaviorSubject<ServiceModel[] | undefined>(undefined);
    private servicesCache$: Observable<ServiceModel[]> | undefined;
    private readonly currentService$: Observable<ServiceModel | undefined>;
    private currentServiceSource = new BehaviorSubject<ServiceModel | undefined>(undefined);
    private currentServiceCache$: Observable<ServiceModel> | undefined;
    private currentServiceId = '';

    constructor(
        private http: HttpClient,
        private messages: MessagesService,
    ) {
        this.currentService$ = this.currentServiceSource.asObservable();
        this.services$ = this.servicesSource.asObservable();
    }

    getServicesList(): Observable<ServiceModel[] | undefined> {
        return this.loadServices().pipe(
            mergeMap(() => this.services$),
        );
    }

    createService(serviceModel: Partial<ServiceModel>): Observable<ServiceModel> {
        return this.http.post<ServiceModel>(`${environment.apiUrl}/services`, serviceModel)
            .pipe(
                tap((updatedService) => {
                    this.reloadServices();
                    this.currentServiceSource.next(updatedService);
                })
            );
    }

    editService(serviceId: string, serviceModel: Partial<ServiceModel>): Observable<ServiceModel> {
        return this.http.put<ServiceModel>(`${environment.apiUrl}/services/${serviceId}`, serviceModel)
            .pipe(
                tap((updatedService) => {
                    this.reloadServices();
                    this.currentServiceSource.next(updatedService);
                })
            );
    }

    deleteService(serviceModel: ServiceModel): Observable<ServiceModel> {
        const serviceId = `${serviceModel.ns}:${serviceModel.name}`;
        return this.http.delete<ServiceModel>(`${environment.apiUrl}/services/${serviceId}`)
            .pipe(
                tap((updatedService) => {
                    this.reloadServices();
                    this.currentServiceSource.next(updatedService);
                })
            );
    }

    getService(serviceId: string): Observable<ServiceModel | undefined> {
        return this.loadService(serviceId).pipe(
            mergeMap(() => {
                return this.currentService$;
            }),
        );
    }

    getCurrentService(): Observable<ServiceModel | undefined> {
        return this.currentService$;
    }

    getCurrentServiceId(): string {
        return this.currentServiceId;
    }

    getCurrentServiceInstancesNames(): Observable<ServiceInstancesNames | undefined> {
        return this.currentService$.pipe(
            filter((v) => v !== undefined),
            map((data) => new Map(data?.instances?.map((v) => [v.name, `${v.name} (${v.role})`])))
        );
    }

    private loadService(serviceId: string): Observable<ServiceModel> {
        if (serviceId !== this.currentServiceId) {
            this.currentServiceCache$ = undefined;
            this.currentServiceId = serviceId;
        }

        if (!this.currentServiceCache$) {
            this.currentServiceCache$ = this.http.get<ServiceModel>(`${environment.apiUrl}/services/${serviceId}`)
                .pipe(
                    publishLast(),
                    refCount(),
                    catchError((err) => {
                        this.messages.error(err);
                        this.currentServiceCache$ = undefined;
                        throw err;
                    }),
                );
        }
        return this.currentServiceCache$.pipe(
            tap((serviceModel) => {
                this.currentServiceSource.next(serviceModel);
                return serviceModel;
            })
        );
    }

    private loadServices(): Observable<ServiceModel[]> {
        if (!this.servicesCache$) {
            this.servicesCache$ = this.http.get<ServiceModel[]>(`${environment.apiUrl}/services`)
                .pipe(
                    publishLast(),
                    refCount(),
                    catchError((err) => {
                        this.messages.error(err);
                        this.servicesCache$ = undefined;
                        throw err;
                    }),
                );
        }
        return this.servicesCache$.pipe(
            tap((serviceModels) => {
                this.servicesSource.next(serviceModels);
                return serviceModels;
            })
        );
    }

    private reloadServices(): void {
        this.servicesCache$ = undefined;
        this.loadServices().subscribe();
    }
}
