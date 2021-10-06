import { Pipe, PipeTransform } from '@angular/core';

// eslint-disable-next-line
export function MockPipeFactory(name: string): any {
    // eslint-disable-next-line
    @Pipe({ name })
    class MockPipe implements PipeTransform {
        transform(value: any): any {
            return `${value}|${name}`;
        }
    }

    return MockPipe;
}
