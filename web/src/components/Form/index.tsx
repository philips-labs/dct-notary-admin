import React, { FC } from 'react';
import { Box, FormField, FormFieldProps, Text } from 'grommet';

export const FormFieldLabel: FC<FormFieldProps> = ({ required, label, name, ...rest }) => (
  <FormField
    name={name}
    label={
      required ? (
        <Box direction="row">
          <Text>{label}</Text>
          <Text color="status-critical">*</Text>
        </Box>
      ) : (
        label
      )
    }
    required={required}
    {...rest}
  />
);
