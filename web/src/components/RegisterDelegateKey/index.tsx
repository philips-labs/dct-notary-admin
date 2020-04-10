import React, { FC } from 'react';
import { RouteComponentProps } from 'react-router-dom';
import { Form, Field, Fields, required } from '..';

type TParams = { targetId: string };

export const RegisterDelegationKey: FC<RouteComponentProps<TParams>> = ({ match }) => {
  const { targetId } = match.params;
  const fields: Fields = {
    delegationName: {
      id: 'delegationName',
      label: 'Name:',
      validator: { rule: required },
    },
    delegationPublicKey: {
      id: 'delegationPublicKey',
      label: 'Public Key:',
      editor: 'multilinetextbox',
      validator: { rule: required },
    },
  };

  return (
    <Form
      action={`/api/targets/${targetId}/delegations`}
      fields={fields}
      render={() => (
        <>
          <div className="row">
            <p>Please paste your public signing key in the form below.</p>
            <pre>cat ~/.docker/trust/your-key.pub | pbcopy</pre>
          </div>
          <div className="row">
            <Field {...fields.delegationName} />
          </div>
          <div className="row">
            <Field {...fields.delegationPublicKey} />
          </div>
        </>
      )}
    />
  );
};
