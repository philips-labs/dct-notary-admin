import React, { FC } from 'react';
import { RouteComponentProps } from 'react-router-dom';
import { Form, Field, Fields, required } from '..';

type TParams = { targetId: string };

export const CreateTarget: FC<RouteComponentProps<TParams>> = () => {
  const fields: Fields = {
    gun: {
      id: 'gun',
      label: 'GUN:',
      validator: { rule: required },
    },
  };

  return (
    <Form action={`/api/targets`} fields={fields}>
      <div className="row">
        <p>E.g.</p>
        <code>
          <pre>localhost:5000/dct-notary-admin</pre>
          <pre>docker.io/philipssoftware/openjdk</pre>
        </code>
      </div>
      <div className="row">
        <Field {...fields.gun} />
      </div>
    </Form>
  );
};
