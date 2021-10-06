import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { ServiceStatusNamePipe } from './service-status-name.pipe';
import { ServiceVersionPipe } from './service-version.pipe';

const pipes = [
    ServiceStatusNamePipe,
    ServiceVersionPipe,
];

@NgModule({
    declarations: pipes,
    imports: [
        CommonModule,
    ],
    exports: pipes
})
export class PipesModule { }
