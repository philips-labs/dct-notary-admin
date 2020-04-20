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
    <Form action={`/api/targets/${targetId}/delegations`} fields={fields}>
      <div className="row">
        <p>First ensure you have a signing key or create a signing key.</p>
        <code>
          <pre>docker trust key generate marcofranssen --dir ~/.docker/trust</pre>
        </code>
        <p>Copy the contents of your public key to the clipboard.</p>
        <pre>cat ~/.docker/trust/marcofranssen.pub | pbcopy</pre>
      </div>
      <div className="row">
        <Field {...fields.delegationName} />
      </div>
      <div className="row">
        <Field {...fields.delegationPublicKey} />
      </div>
    </Form>
  );
};
