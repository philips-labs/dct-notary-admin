import { FC, FormEvent, useState, useContext } from 'react';
import { Form, TextInput, TextArea } from 'grommet';
import axios from 'axios';
import { DelegationContext } from './DelegationContext';
import { FormFieldLabel } from '../Form';

interface TargetParams {
  targetId: string;
}
type RegisterDelegationKeyState = {
  delegationName: string;
  delegationPublicKey: string;
  errorMessage: string;
};

const defaultFormValue = {
  delegationName: '',
  delegationPublicKey: '',
  errorMessage: '',
};

export const RegisterDelegationKey: FC<TargetParams> = ({ targetId }) => {
  const [value, setValue] = useState<RegisterDelegationKeyState>(defaultFormValue);
  const { refresh } = useContext(DelegationContext);
  const delegationNameMsg = 'may only contain a-z and _';
  const submitForm = async (event: FormEvent) => {
    event.preventDefault();
    try {
      await axios.post(`/api/targets/${targetId}/delegations`, JSON.stringify(value), {
        headers: new Headers({
          'Content-Type': 'application/json',
          Accept: 'application/json',
        }),
      });
      setValue(defaultFormValue);
      refresh();
    } catch (e) {
      const response = e.response;
      const errorMessage = `${response.data.status} ${response.data.error}`;
      setValue({ ...value, errorMessage });
      console.log(value);
    }
  };

  return (
    <Form
      value={value}
      onChange={(event: any) => {
        setValue(event as RegisterDelegationKeyState);
      }}
      onSubmit={submitForm}
      validate="blur"
    >
      <p className="text-gray-600 my-2">
        First ensure you have a signing key or create a signing key.
      </p>
      <FormFieldLabel
        label="Name"
        name="delegationName"
        required
        help={delegationNameMsg}
        validate={[{ regexp: /^[a-z_]*$/, message: delegationNameMsg, status: 'error' }]}
      >
        <TextInput name="delegationName" placeholder="marcofranssen" required />
      </FormFieldLabel>
      <FormFieldLabel
        name="delegationPublicKey"
        help="cat ~/.docker/trust/marcofranssen.pub | pbcopy"
        label="Public Key"
        required
      >
        <TextArea name="delegationPublicKey" required />
      </FormFieldLabel>
      {value.errorMessage && <p className="text-sm text-red-500 p-1">{value.errorMessage}</p>}
      <div className="flex flex-row-reverse">
        <button
          className="bg-blue-600 text-white p-2 px-5 hover:bg-blue-700 rounded-3xl font-semibold"
          type="submit"
        >
          Submit
        </button>
      </div>
    </Form>
  );
};
