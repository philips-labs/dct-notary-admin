import { FC, FormEvent, useState, useContext, ChangeEvent } from 'react';
import axios from 'axios';
import { DelegationContext } from './DelegationContext';
import { FormTextInput, FormTextArea } from '../Form';

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
      const { errorMessage, ...requestBody } = value;
      await axios.post(`/api/targets/${targetId}/delegations`, JSON.stringify(requestBody), {
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json',
        },
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
    <form onSubmit={submitForm} className="antialiased">
      <p className="text-gray-600 my-2">
        First ensure you have a signing key or create a signing key.
      </p>
      <FormTextInput
        label="Name"
        name="delegationName"
        required
        placeholder="marcofranssen"
        help={delegationNameMsg}
        value={value.delegationName}
        onChange={(event: ChangeEvent<HTMLInputElement>) => {
          setValue((prevValue) => ({ ...prevValue, delegationName: event.target.value }));
        }}
        className="mb-3"
      />
      <FormTextArea
        label="Public Key"
        name="delegationPublicKey"
        help="cat ~/.docker/trust/marcofranssen.pub | pbcopy"
        value={value.delegationPublicKey}
        placeholder="-----BEGIN PUBLIC KEY-----
        role: marcofranssen

        MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEmI6bhcF0aqKobYIgBD/wHg/vhjW2
        E+C9PEdgfom/x+XxcrFLxvPz1jl7sH8yj315Tr3C5dcE9GhDDlNyJcNC/g==
        -----END PUBLIC KEY-----"
        onChange={(event: ChangeEvent<HTMLTextAreaElement>) => {
          setValue((prevValue) => ({ ...prevValue, delegationPublicKey: event.target.value }));
        }}
        required
      />
      {value.errorMessage && <p className="text-sm text-red-500 p-1">{value.errorMessage}</p>}
      <div className="flex flex-row-reverse">
        <button
          className="bg-blue-600 text-white p-2 px-5 hover:bg-blue-700 rounded-3xl font-semibold focus:outline-none"
          type="submit"
        >
          Submit
        </button>
      </div>
    </form>
  );
};
