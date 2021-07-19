import { functionB } from '@simple-workspace/b'

export function functionA () {
  return {
    a: 'a',
    b: functionB()
  }
}
