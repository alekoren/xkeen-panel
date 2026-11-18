import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Logo } from '@/components/logo'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuth } from '@/hooks/useAuth'

export function LoginPage() {
 const { login } = useAuth()
 const [username,setUsername] = useState('')
 const [password,setPassword] = useState('')
 const mutation = useMutation({ mutationFn: () => login(username,password,'') })
 return <div className='min-h-screen flex items-center justify-center p-4'>
  <div aria-hidden className='bg-page fixed inset-0 -z-10' />
  <div className='w-full max-w-md'><div className='mb-5 flex justify-center'><Logo size={56} /></div>
   <Card><CardHeader className='text-center'><CardTitle className='text-2xl'>XKeen</CardTitle><CardDescription>Вход по логину и паролю</CardDescription></CardHeader>
    <CardContent><form className='space-y-4' onSubmit={event => { event.preventDefault(); mutation.mutate() }}>
     <Label htmlFor='login-username'>Логин</Label><Input id='login-username' autoComplete='username' value={username} onChange={e => setUsername(e.target.value)} required autoFocus />
     <Label htmlFor='login-password'>Пароль</Label><Input id='login-password' type='password' autoComplete='current-password' value={password} onChange={e => setPassword(e.target.value)} required />
     {mutation.error && <p role='alert' className='text-destructive'>{mutation.error.message}</p>}
     <Button type='submit' disabled={mutation.isPending} className='w-full'>{mutation.isPending ? 'Вход…' : 'Войти'}</Button>
    </form></CardContent>
   </Card>
  </div>
 </div>
}
