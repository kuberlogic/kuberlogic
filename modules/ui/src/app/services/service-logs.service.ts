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
