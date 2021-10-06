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
