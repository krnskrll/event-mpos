/** Indonesian Rupiah display: "Rp" prefix, "." thousands grouping, no decimals. */
export function rupiah(amount: number): string {
  return "Rp " + Math.round(amount).toLocaleString("id-ID");
}
