import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from '@environments/environment';
import { ServiceBackupModel } from '@models/service-backup.model';
import { MessagesService } from '@services/messages.service';
import { BehaviorSubject, Observable, of } from 'rxjs';
import { catchError, mergeMap, publishLast, refCount, tap } from 'rxjs/operators';

@Injectable({
    providedIn: 'root'
})
export class ServiceBackupsService {

    private readonly data$: Observable<ServiceBackupModel[] | undefined>;
    private dataSource = new BehaviorSubject<ServiceBackupModel[] | undefined>(undefined);
    private cache$: Observable<ServiceBackupModel[]> | undefined;
    private currentServiceId = '';
    private notFoundRegexp = new RegExp(/^secrets ".*" not found$/);

    constructor(
        private http: HttpClient,
        private messages: MessagesService,
    ) {
        this.data$ = this.dataSource.asObservable();
    }

    getList(serviceId: string): Observable<ServiceBackupModel[] | undefined> {
        return this.load(serviceId).pipe(
            mergeMap(() => this.data$),
        );
    }

    private load(serviceId: string): Observable<ServiceBackupModel[]> {
        if (serviceId !== this.currentServiceId) {
            this.cache$ = undefined;
            this.currentServiceId = serviceId;
        }

        if (!this.cache$) {
            this.cache$ = this.http.get<ServiceBackupModel[]>(`${environment.apiUrl}/services/${serviceId}/backups`)
                .pipe(
                    publishLast(),
                    refCount(),
                    catchError((err) => {
                        this.cache$ = undefined;
                        if (this.notFoundRegexp.test(err)) {
                            return of([]);
                        } else {
                            this.messages.error(err);
                            throw err;
                        }
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
}
