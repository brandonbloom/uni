import { functionD } from '@simple-workspace/d'
import { functionE } from '@simple-workspace/e'

export function functionC () {
  return {
    c: 'c',
    d: functionD(),
    e: functionE()
  }
}
