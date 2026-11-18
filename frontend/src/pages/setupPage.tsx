import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api, setToken } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export function SetupPage() {
 const [username,setUsername] = useState('')
 const [password,setPassword] = useState('')
 const [repeat,setRepeat] = useState('')
 const setup = useMutation({
  mutationFn: () => api.post<{token: string}>('/api/auth/setup',{username,password}),
  onSuccess: data => { setToken(data.token); window.location.assign('/') },
 })
 return <div className='min-h-screen flex items-center justify-center p-4'><div className='w-full max-w-md'>
  <Card><CardHeader><CardTitle>Настройка XKeen Panel</CardTitle></CardHeader><CardContent>
   <form className='space-y-4' onSubmit={event => { event.preventDefault(); if (password === repeat) setup.mutate() }}>
    <Label htmlFor='setup-user'>Логин</Label><Input id='setup-user' required autoComplete='username' maxLength={100} value={username} onChange={e => setUsername(e.target.value)} />
    <Label htmlFor='setup-password'>Пароль (от 8 до 72 байт)</Label><Input id='setup-password' required minLength={8} type='password' autoComplete='new-password' value={password} onChange={e => setPassword(e.target.value)} />
    <Label htmlFor='setup-repeat'>Повторите пароль</Label><Input id='setup-repeat' required type='password' autoComplete='new-password' value={repeat} onChange={e => setRepeat(e.target.value)} />
    {repeat && password !== repeat && <p role='alert'>Пароли не совпадают</p>}
    {setup.error && <p role='alert' className='text-destructive'>{setup.error.message}</p>}
    <Button type='submit' disabled={setup.isPending || password !== repeat}>{setup.isPending ? 'Сохраняю…' : 'Создать пользователя'}</Button>
   </form>
  </CardContent></Card>
 </div></div>
}
