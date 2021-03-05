import { FC } from 'react';
import { FormField, FormFieldProps } from 'grommet';

export const FormFieldLabel: FC<FormFieldProps> = ({ required, label, name, ...rest }) => (
  <FormField
    name={name}
    label={
      required ? (
        <div className="flex flex-row">
          <span>{label}</span>
          <span className="text-red-500 ml-1">*</span>
        </div>
      ) : (
        label
      )
    }
    required={required}
    {...rest}
  />
);
