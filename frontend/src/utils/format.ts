export const formatCount = (value: number): string => {
  if (value >= 10000) {
    const count = Math.round((value / 10000) * 10) / 10
    return `${count}万`
  }
  return String(value)
}
