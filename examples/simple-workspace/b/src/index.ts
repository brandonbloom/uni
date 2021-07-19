import { functionC } from '@simple-workspace/c'
import { functionD } from '@simple-workspace/d'

export function functionB () {
  return {
    b: 'b',
    c: functionC(),
    d: functionD()
  }
}
