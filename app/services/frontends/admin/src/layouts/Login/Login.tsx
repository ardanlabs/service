'use client'

import * as React from 'react'
import Container from '@mui/material/Container'
import Typography from '@mui/material/Typography'
import Box from '@mui/material/Box'
import Copyright from '@/components/CopyRight/Copyright'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import PasswordTextField from '@/components/PasswordTextField/PasswordTextField'

export default function Login() {
  const [formData, setFormData] = React.useState({ username: '', password: '' })

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
            <PasswordTextField
              label="Password"
              name="password"
              handleOnChange={handleChange}
            />
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
