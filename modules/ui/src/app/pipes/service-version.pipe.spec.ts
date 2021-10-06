import { ServiceModel, ServiceModelStatus, ServiceModelType } from '@models/service.model';
import { ServiceVersionPipe } from './service-version.pipe';

const serviceModel: ServiceModel = {
    type: ServiceModelType.POSTGRES,
    name: 'PostgreSql prod',
    status: ServiceModelStatus.FAILED,
    masters: 2,
    replicas: 5,
    created_at: '2021-02-09T10:56:12.115Z',
};

describe('ServiceVersionPipe', () => {
    const pipe = new ServiceVersionPipe();

    it('create an instance', () => {
        expect(pipe).toBeTruthy();
    });

    it('should return "PostgreSQL 12.1.5" when type is postgre', () => {
        const result = pipe.transform(serviceModel);
        expect(result).toEqual('PostgreSQL 12.1.5');
    });

    it('should return "MySQL 5.7.31" when type is mysql', () => {
        const result = pipe.transform({
            ...serviceModel,
            type: ServiceModelType.MYSQL
        });
        expect(result).toEqual('MySQL 5.7.31');
    });
});
