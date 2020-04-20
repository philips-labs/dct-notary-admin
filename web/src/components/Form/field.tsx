import React, { FC, FormEvent, useContext } from 'react';
import { FormContext, FormValues, FormErrors } from '.';

export type Editor = 'textbox' | 'multilinetextbox' | 'dropdown';

export const required = (values: FormValues, fieldName: string): string =>
  values[fieldName] === undefined || values[fieldName] === null || values[fieldName] === ''
    ? 'This field is required be populated'
    : '';

export interface Validator {
  rule: (values: FormValues, fieldName: string, args: any) => string;
  args?: any;
}

export interface FieldProps {
  id: string;
  label?: string;
  editor?: Editor;
  options?: string[];
  value?: any;
  validator?: Validator;
}

// const useInput = (initialValue) => {
//   const [value, setValue] = useState(initialValue);

//   return {
//     value,
//     setValue,
//     reset: () => setValue(''),
//     bind: {
//       value,
//       onChange: (event: FormEvent<HTMLInputElement>) => {
//         setValue(event.target.value);
//       },
//     },
//   };
// };

export const Field: FC<FieldProps> = ({ id, label, editor, options, value }) => {
  const context = useContext(FormContext);
  const getError = (errors?: FormErrors): string => (errors ? errors[id] : '');
  const getEditorStyle = (errors?: FormErrors): any =>
    getError(errors) ? { borderColor: 'red' } : {};

  return (
    <>
      {label && (
        <div className="col-25">
          <label htmlFor={id}>{label}</label>
        </div>
      )}
      <div className="col-75">
        {editor!.toLowerCase() === 'textbox' && (
          <input
            id={id}
            type="text"
            value={value}
            onChange={(e: FormEvent<HTMLInputElement>) =>
              context?.setValues({ [id]: e.currentTarget.value })
            }
            style={getEditorStyle(context?.errors)}
            onBlur={() => context?.validate(id)}
          />
        )}

        {editor!.toLowerCase() === 'multilinetextbox' && (
          <textarea
            id={id}
            value={value}
            onChange={(e: React.FormEvent<HTMLTextAreaElement>) =>
              context?.setValues({ [id]: e.currentTarget.value })
            }
            onBlur={() => context?.validate(id)}
            style={getEditorStyle(context?.errors)}
            className="form-control"
          />
        )}

        {editor!.toLowerCase() === 'dropdown' && (
          <select
            id={id}
            name={id}
            value={value}
            onChange={(e: React.FormEvent<HTMLSelectElement>) =>
              context?.setValues({ [id]: e.currentTarget.value })
            }
            onBlur={() => context?.validate(id)}
            style={getEditorStyle(context?.errors)}
            className="form-control"
          >
            {options &&
              options.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
          </select>
        )}
        {getError(context?.errors) && (
          <div style={{ color: 'red', fontSize: '80%' }}>
            <p>{getError(context?.errors)}</p>
          </div>
        )}
      </div>
    </>
  );
};

Field.defaultProps = {
  editor: 'textbox',
};
