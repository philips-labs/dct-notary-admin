import { FC } from 'react';
import { Route } from 'react-router-dom';
import { Delegations, Targets } from '../../components';

export const TargetsPage: FC = () => {
  return (
    <div className="overflow-auto grid grid-cols-1 lg:grid-cols-2 gap-10">
      <div className="p-3">
        <h2 className="text-4xl tracking-tight font-bold leading-normal">Targets</h2>
        <Targets />
      </div>
      <div className="p-3">
        <h2 className="text-4xl tracking-tight font-bold leading-normal">Delegations</h2>
        <Route path="/:targetId">
          <Delegations />
        </Route>
      </div>
    </div>
  );
};
