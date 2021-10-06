import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { HttpClientModule } from '@angular/common/http';
import { MatButtonModule } from '@angular/material/button';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { InterceptorsModule } from '@app/interceptors/interceptors.module';
import { ConfirmDialogModule } from '@ui/confirm-dialog/confirm-dialog.module';
import { HeaderModule } from '@ui/header/header.module';
import { TimeagoModule } from 'ngx-timeago';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';

@NgModule({
    declarations: [
        AppComponent
    ],
    imports: [
        BrowserModule,
        AppRoutingModule,
        BrowserAnimationsModule,
        MatSnackBarModule,
        MatButtonModule,
        HeaderModule,
        HttpClientModule,
        TimeagoModule.forRoot(),
        InterceptorsModule,
        ConfirmDialogModule,
    ],
    providers: [MatButtonModule],
    bootstrap: [AppComponent]
})
export class AppModule { }
