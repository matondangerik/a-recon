# Transaction Reconciliation Service

This repository contains my solution for the Amartha reconciliation exercise.

It compares system transactions against bank statement transactions and reports:

- total processed transactions
- total matched transactions
- total unmatched transactions
- unmatched system transactions
- unmatched bank transactions grouped by bank
- total discrepancy across matched transactions

## Project Layout

- `cmd/cli`
  - CLI entrypoint
- `internal/usecase/reconciliation`
  - parsing, matching, and summary generation
- `pkg/csv`
  - CSV loader wrapper
- `pkg/errgroup`
  - thin `errgroup` wrapper
- `doc`
  - sample CSV inputs and example outputs

## Assumptions

The exercise leaves a few things open, so I made these choices:

- reconciliation is done per bank
- `bank_name` is added to the system CSV for routing to the correct bank file
- the CLI derives bank names from the bank file basename without extension
  - `bank-a.csv` -> `bank-a`
- all time values are assumed to be UTC
- system amounts are non-negative and direction comes from `type`
- bank amounts use sign for direction
  - negative = debit
  - positive = credit
- bank statement `date` is treated as a date, not a full timestamp
- bank statement `date` is expected in `YYYY-MM-DD` format
- CSV inputs are assumed valid
- unknown banks in the system file are ignored

I also added tolerance as a practical matching feature. That is my own implementation choice, not something explicitly required by the prompt.

- `--tolerance` controls day tolerance
- `--amount-tolerance` controls amount tolerance

If you want strict matching, set both to `0`.

## CSV Contract

### System CSV

```csv
trx_id,amount,type,transaction_time,bank_name
```

```csv
trx-credit-match,100,CREDIT,2026-01-10T10:00:00Z,bank-a
trx-debit-match,100,DEBIT,2026-01-10T11:00:00Z,bank-b
```

### Bank CSV

```csv
unique_identifier,amount,date
```

```csv
bank-credit-match,103,2026-01-10
bank-debit-match,-100,2026-01-10
```

## Matching Rules

At a high level:

1. parse system and bank CSVs
2. filter by date range
3. group system rows by bank
4. compare each system row only against the same bank
5. accept a match only if direction, date window, and amount tolerance all pass
6. build the summary and unmatched sections

`Total Transactions` in this implementation means reconciliation result rows:

- matched rows
- unmatched system rows
- unmatched bank rows

## Running The CLI

```bash
go run ./cmd/cli \
  --system doc/system_transactions.csv \
  --banks doc/bank-a.csv,doc/bank-b.csv \
  --start 2026-01-10 \
  --end 2026-01-17 \
  --tolerance 1 \
  --amount-tolerance 5 \
  --output doc/cli_output.txt
```

Important flags:

- `--system`: system CSV path
- `--banks`: comma-separated bank CSV paths
- `--start`: start date, `YYYY-MM-DD`
- `--end`: end date, `YYYY-MM-DD`
- `--tolerance`: day tolerance, `0` disables it
- `--amount-tolerance`: amount tolerance, `0` means exact amount
- `--output`: optional output file path; output is still printed to stdout

## Sample Data

The `doc` folder contains one combined system file and two combined bank files:

- `doc/system_transactions.csv`
- `doc/bank-a.csv`
- `doc/bank-b.csv`

It also contains example outputs:

- `doc/cli_output.txt`
- `doc/cli_output_no_tolerance.txt`

The sample data includes:

- credit match
- debit match
- unmatched system row
- unmatched bank row
- tolerance boundary cases
- duplicate candidates
- zero amount

## Running Tests

```bash
go test ./...
```
