from pathlib import Path

path = Path("scripts/stage375_apply.py")
text = path.read_text()
start_marker = "# Existing transaction form exposes all canonical types and exact shape/null semantics."
end_marker = "# Portfolio page loads cash-flow projection"
start = text.find(start_marker)
end = text.find(end_marker)
if start < 0 or end < 0 or end <= start:
    raise SystemExit("unable to locate transaction-form patch section")

replacement = """# Existing transaction form exposes all canonical types and exact shape/null semantics.
path = \"frontend-next/src/features/portfolio/components/AddTransactionForm.tsx\"
text = read(path)
text = text.replace(
    'const transactionTypes: TransactionType[] = [\"BUY\", \"SELL\", \"DEPOSIT\", \"WITHDRAWAL\"];',
    'const transactionTypes: TransactionType[] = [\"BUY\", \"SELL\", \"DIVIDEND\", \"COUPON\", \"FEE\", \"TAX\", \"DEPOSIT\", \"WITHDRAWAL\"];',
    1,
)
text = text.replace(
    '  const isAssetIncome = transactionType === \"DIVIDEND\" || transactionType === \"COUPON\";\\n  const isCashFlow = transactionType === \"DEPOSIT\" || transactionType === \"WITHDRAWAL\";',
    '  const isAssetIncome = transactionType === \"DIVIDEND\" || transactionType === \"COUPON\";\\n  const isExpense = transactionType === \"FEE\" || transactionType === \"TAX\";\\n  const isCashFlow = transactionType === \"DEPOSIT\" || transactionType === \"WITHDRAWAL\";',
    1,
)
old = '''  function buildPayload(normalizedNote: string): CreateTransactionPayload {
    const assetTicker = isCashFlow ? null : ticker.trim().toUpperCase();
    return {
      transactionType,
      ticker: assetTicker,
      quantity: isTrade || isAssetIncome ? quantity : null,
      unitPrice: isTrade ? { amount: unitPrice, currency: \"RUB\" } : null,
      grossAmount: isTrade ? null : { amount: grossAmount, currency: \"RUB\" },
      commission: { amount: commission, currency: \"RUB\" },
      tax: { amount: tax, currency: \"RUB\" },
      tradeDate,
      settlementDate: settlementDate.trim() === \"\" ? null : settlementDate,
      note: normalizedNote === \"\" ? null : normalizedNote,
    };
  }
'''
new = '''  function buildPayload(normalizedNote: string): CreateTransactionPayload {
    const assetTicker = isTrade || isAssetIncome ? ticker.trim().toUpperCase() : null;
    return {
      transactionType,
      ticker: assetTicker,
      quantity: isTrade ? quantity : isAssetIncome && quantity.trim() !== \"\" ? quantity : null,
      unitPrice: isTrade ? { amount: unitPrice, currency: \"RUB\" } : null,
      grossAmount: isTrade ? null : { amount: grossAmount, currency: \"RUB\" },
      commission: { amount: isExpense || isCashFlow ? \"0.00000000\" : commission, currency: \"RUB\" },
      tax: { amount: isExpense || isCashFlow ? \"0.00000000\" : tax, currency: \"RUB\" },
      tradeDate,
      settlementDate: settlementDate.trim() === \"\" ? null : settlementDate,
      note: normalizedNote === \"\" ? null : normalizedNote,
    };
  }
'''
if old in text:
    text = text.replace(old, new, 1)
elif new not in text:
    raise SystemExit(\"missing AddTransactionForm buildPayload anchor\")
text = text.replace(
    '          transactions, recalculates snapshots, and returns canonical results. Stage 3.71 exposes\\n          manual BUY and SELL while broker-import SELL remains intentionally unavailable.',
    '          transactions, recalculates snapshots, and returns canonical results. Stage 3.75 exposes\\n          manual trades, income, expenses, deposits and withdrawals while broker-import SELL remains intentionally unavailable.',
    1,
)
text = text.replace('{!isCashFlow ? (\\n        <label>', '{isTrade || isAssetIncome ? (\\n        <label>', 1)
text = text.replace('<input value={quantity} required inputMode=\"decimal\" onChange={(event) => setQuantity(event.target.value)} />', '<input value={quantity} required={isTrade} inputMode=\"decimal\" onChange={(event) => setQuantity(event.target.value)} />', 1)
old = '''      <label>
        Commission
        <input value={commission} required inputMode=\"decimal\" onChange={(event) => setCommission(event.target.value)} />
      </label>

      <label>
        Tax
        <input value={tax} required inputMode=\"decimal\" onChange={(event) => setTax(event.target.value)} />
      </label>
'''
new = '''      {!isExpense && !isCashFlow ? (
        <>
          <label>
            Commission
            <input value={commission} required inputMode=\"decimal\" onChange={(event) => setCommission(event.target.value)} />
          </label>

          <label>
            Tax
            <input value={tax} required inputMode=\"decimal\" onChange={(event) => setTax(event.target.value)} />
          </label>
        </>
      ) : null}
'''
if old in text:
    text = text.replace(old, new, 1)
elif new not in text:
    raise SystemExit(\"missing AddTransactionForm commission/tax anchor\")
write(path, text)

"""

path.write_text(text[:start] + replacement + text[end:])
