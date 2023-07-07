'use client'

import * as React from 'react'
import Container from '@mui/material/Container'
import Typography from '@mui/material/Typography'
import Box from '@mui/material/Box'
import Copyright from '@/components/CopyRight/Copyright'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import FilledInput from '@mui/material/FilledInput'
import FormControl from '@mui/material/FormControl'
import IconButton from '@mui/material/IconButton'
import InputAdornment from '@mui/material/InputAdornment'
import InputLabel from '@mui/material/InputLabel'
import VisibilityOff from '@mui/icons-material/VisibilityOff'
import Visibility from '@mui/icons-material/Visibility'

export default function Login() {
  const [formData, setFormData] = React.useState({ username: '', password: '' })
  const [showPassword, setShowPassword] = React.useState(false)

  const handleClickShowPassword = () => setShowPassword((show) => !show)

  const handleMouseDownPassword = (
    event: React.MouseEvent<HTMLButtonElement>,
  ) => {
    event.preventDefault()
  }

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = event.target
    setFormData((prevFormData) => ({ ...prevFormData, [name]: value }))
  }

  const handleSubmit = async () => {
    try {
      const response = await fetch(
        `${process.env.baseAPIUrl}/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1`,
      )
      if (!response.ok) {
        throw new Error('Network response was not OK')
      }
    } catch (error) {
      console.error(
        'There has been a problem with your fetch operation:',
        error,
      )
    }
  }

  React.useEffect(() => {}, [formData])
  return (
    <Container maxWidth="xl" disableGutters>
      <Box sx={{ pt: 4 }}>
        <form
          onSubmit={handleSubmit}
          style={{
            display: 'flex',
            alignContent: 'center',
            justifyContent: 'center',
          }}
        >
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignContent: 'center',
              justifyContent: 'center',
              width: '400px',
              maxWidth: '80%',
              mx: 4,
              mb: 6,
              p: 4,
              borderRadius: '4px',
              backgroundColor: '#151420',
            }}
          >
            <Typography variant="subtitle1" align="center" color="white">
              {'Welcome back!'}
            </Typography>
            <TextField
              required
              id="filled-required"
              label="Username"
              name="username"
              variant="filled"
              sx={{
                my: 2,
                backgroundColor: '#FFFFFF',
                borderRadius: '4px',
              }}
              onChange={handleChange}
            />
            <FormControl
              sx={{
                my: 2,
                backgroundColor: '#FFFFFF',
                borderRadius: '4px',
              }}
              variant="filled"
            >
              <InputLabel htmlFor="outlined-adornment-password">
                Password
              </InputLabel>
              <FilledInput
                id="outlined-adornment-password"
                type={showPassword ? 'text' : 'password'}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={handleClickShowPassword}
                      onMouseDown={handleMouseDownPassword}
                      edge="end"
                    >
                      {showPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                }
                label="Password"
              />
            </FormControl>
            <Button
              type="submit"
              variant="contained"
              color="primary"
              sx={{ my: 2 }}
            >
              Login
            </Button>
          </Box>
        </form>
        <Box sx={{ justifySelf: 'end' }}>
          <Copyright />
        </Box>
      </Box>
    </Container>
  )
}
