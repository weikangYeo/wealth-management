import { Component } from '@angular/core';

@Component({
  selector: 'app-home-page',
  standalone: true,
  imports: [],
  template: `
    <div class="p-6">
      <h2 class="text-3xl font-semibold text-gray-800">Welcome to Wealth Management</h2>
      <p class="mt-4 text-gray-600">This is your personalized dashboard to manage your gold and funds.</p>
      <p class="mt-2 text-gray-600">Use the navigation on the left to explore different sections.</p>
    </div>
  `,
  styles: []
})
export class HomePage { }