import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { environment } from '@environments/environment';
import { ServiceBackupConfigModel } from '@models/service-backup-config.model';
import { MessagesService } from '@services/messages.service';
import { BehaviorSubject, EMPTY, Observable } from 'rxjs';
import { catchError, mergeMap, publishLast, refCount, tap } from 'rxjs/operators';

@Injectable({
    providedIn: 'root'
})
export class BackupConfigService {

    private readonly currentBackupConfig$: Observable<ServiceBackupConfigModel | undefined>;
    private currentBackupConfigSource = new BehaviorSubject<ServiceBackupConfigModel | undefined>(undefined);
    private currentBackupConfigCache$: Observable<ServiceBackupConfigModel | undefined> | undefined;
    private currentServiceId = '';

    constructor(
        private http: HttpClient,
        private messages: MessagesService,
    ) {
        this.currentBackupConfig$ = this.currentBackupConfigSource.asObservable();
    }

    getBackupConfig(serviceId: string): Observable<ServiceBackupConfigModel | undefined> {
        return this.loadBackupConfig(serviceId).pipe(
            mergeMap(() => {
                return this.currentBackupConfig$;
            }),
        );
    }

    getCurrentBackupConfig(): Observable<ServiceBackupConfigModel | undefined> {
        return this.currentBackupConfig$;
    }

    createBackupConfig(serviceId: string, config: ServiceBackupConfigModel): Observable<ServiceBackupConfigModel> {
        return this.http
            .post<ServiceBackupConfigModel>(`${environment.apiUrl}/services/${serviceId}/backup-config`, config);
    }

    editBackupConfig(serviceId: string, config: ServiceBackupConfigModel): Observable<ServiceBackupConfigModel> {
        return this.http
            .put<ServiceBackupConfigModel>(`${environment.apiUrl}/services/${serviceId}/backup-config`, config)
            .pipe(
                tap((updatedBackup) => {
                    this.currentBackupConfigSource.next(updatedBackup);
                }),
            );
    }

    private loadBackupConfig(serviceId: string): Observable<ServiceBackupConfigModel | undefined> {
        if (serviceId !== this.currentServiceId) {
            this.currentBackupConfigCache$ = undefined;
            this.currentServiceId = serviceId;
        }

        if (!this.currentBackupConfigCache$) {
            this.currentBackupConfigCache$ = this.http
                .get<ServiceBackupConfigModel | undefined>(`${environment.apiUrl}/services/${serviceId}/backup-config`)
                .pipe(
                    publishLast(),
                    refCount(),
                    catchError((err) => {
                        this.currentBackupConfigCache$ = undefined;

                        // 404 means only that `Automatic backups` is turned off
                        if (err === 'Not Found') {
                            return EMPTY;
                        } else {
                            this.messages.error(err);
                            throw err;
                        }
                    }),
                );
        }
        return this.currentBackupConfigCache$.pipe(
            tap((backup) => {
                this.currentBackupConfigSource.next(backup);
            })
        );
    }
}
