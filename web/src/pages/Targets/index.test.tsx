import { act, render } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { TargetsPage } from '.';

describe('TargetPage', () => {
  it('should render a loading state', async () => {
    await act(async () => {
      const { getByText } = render(
        <MemoryRouter>
          <TargetsPage />
        </MemoryRouter>,
      );
      expect(getByText('Loading...')).toBeInTheDocument();
    });
  });
});
