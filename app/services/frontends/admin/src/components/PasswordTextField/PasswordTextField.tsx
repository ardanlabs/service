'use client'

import * as React from 'react'
import FilledInput from '@mui/material/FilledInput'
import FormControl from '@mui/material/FormControl'
import IconButton from '@mui/material/IconButton'
import InputAdornment from '@mui/material/InputAdornment'
import InputLabel from '@mui/material/InputLabel'
import VisibilityOff from '@mui/icons-material/VisibilityOff'
import Visibility from '@mui/icons-material/Visibility'
import { FormHelperText, SxProps } from '@mui/material'
import { Theme } from '@mui/material'

interface PasswordTextFieldProps {
  label: string
  name: string
  styles?: SxProps<Theme>
  error?: boolean
  helperText?: string
  handleOnChange: (event: React.ChangeEvent<HTMLInputElement>) => void
}

export default function PasswordTextField(props: PasswordTextFieldProps) {
  const { label, handleOnChange, name, styles, error, helperText } = props
  const [showPassword, setShowPassword] = React.useState(false)

  const handleClickShowPassword = () => setShowPassword((show) => !show)

  const handleMouseDownPassword = (
    event: React.MouseEvent<HTMLButtonElement>,
  ) => {
    event.preventDefault()
  }

  return (
    <FormControl
      sx={{
        my: 2,
        backgroundColor: '#FFFFFF',
        borderRadius: '4px',
        ...styles,
      }}
      error={error}
      variant="filled"
    >
      <InputLabel htmlFor="outlined-adornment-password">{label}</InputLabel>
      <FilledInput
        id="outlined-adornment-password"
        name={name}
        type={showPassword ? 'text' : 'password'}
        error={error}
        onChange={handleOnChange}
        endAdornment={
          <InputAdornment position="end">
            <IconButton
              aria-label="toggle password visibility"
              onClick={handleClickShowPassword}
              onMouseDown={handleMouseDownPassword}
              edge="end"
            >
              {showPassword ? <Visibility /> : <VisibilityOff />}
            </IconButton>
          </InputAdornment>
        }
      />
      {error ? (
        <FormHelperText id="my-helper-text">{helperText}</FormHelperText>
      ) : null}
    </FormControl>
  )
}
