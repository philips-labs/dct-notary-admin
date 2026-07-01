import { render, screen } from '@testing-library/react';
import axios from 'axios';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, vi } from 'vitest';
import { TargetsPage } from '.';

vi.mock('axios');

const mockedAxios = vi.mocked(axios, true);

describe('TargetPage', () => {
  beforeEach(() => {
    mockedAxios.get.mockResolvedValue({ data: [] });
  });

  it('should render a loading state', async () => {
    render(
      <MemoryRouter>
        <TargetsPage />
      </MemoryRouter>,
    );

    expect(screen.getByText('Loading…')).toBeInTheDocument();
  });
});
