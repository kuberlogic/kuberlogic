import { Pipe, PipeTransform } from '@angular/core';
import { ServiceModel, ServiceModelType } from '@models/service.model';

@Pipe({
    name: 'klServiceVersion'
})
export class ServiceVersionPipe implements PipeTransform {

    transform(serviceModel: ServiceModel | undefined): string {
        if (!!serviceModel) {
            // @TODO remove hardcoded version when version support is implemented
            return serviceModel.type === ServiceModelType.POSTGRES ? 'PostgreSQL 12.1.5' : 'MySQL 5.7.31';
        }
        return '';
    }

}
