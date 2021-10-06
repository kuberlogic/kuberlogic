import { ChangeDetectionStrategy, Component, EventEmitter, Output } from '@angular/core';

@Component({
    selector: 'kl-create-service-header',
    templateUrl: './create-service-header.component.html',
    styleUrls: ['./create-service-header.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class CreateServiceHeaderComponent {
    @Output() createPressed = new EventEmitter<void>();

    onCreatePressed(): void {
        this.createPressed.emit();
    }

}
