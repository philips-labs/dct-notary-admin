import React, { FC, FormEvent, useState, useContext } from 'react';
import { Box, Text, Form, TextInput, TextArea, Button, Paragraph } from 'grommet';
import axios from 'axios';
import { DelegationContext } from './DelegationContext';
import { FormFieldLabel } from '../Form';

interface TargetParams {
  targetId: string;
}
type RegisterDelegationKey = {
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
  const [value, setValue] = useState<RegisterDelegationKey>(defaultFormValue);
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
        setValue(event as RegisterDelegationKey);
      }}
      onSubmit={submitForm}
      validate="blur"
    >
      <Paragraph fill>First ensure you have a signing key or create a signing key.</Paragraph>
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
      {value.errorMessage && (
        <Box pad={{ horizontal: 'small' }}>
          <Text color="status-error">{value.errorMessage}</Text>
        </Box>
      )}
      <Box direction="row" justify="end" margin={{ top: 'medium' }}>
        <Button type="submit" label="Submit" primary />
      </Box>
    </Form>
  );
};
