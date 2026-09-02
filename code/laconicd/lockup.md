# Lockup Account Usage

* Add a genesis lockup account:

  ```bash
  laconicd genesis add-genesis-lockup-account <account_name> <distribution-json-file> <coin>[,<coin>...]

  # Example
  # laconicd genesis add-genesis-lockup-account lps_lockup distribution.json 1000alps
  ```

  * This adds a `LockupAccount` with given name and balance in the genesis file

  * The lockup account can be queried as shown below once the chain starts

* Query a lockup account:

  ```bash
  laconicd query auth module-account <account_name>

  # Example
  # laconicd query auth module-account lps_lockup
  # account:
  # type: laconic/LockupAccount
  # value:
  #   base_account:
  #     account_number: "1"
  #     address: laconic1mprsxp9jqe0d0lp88fxuccthwgy7tqgt5x9y65
  #   distribution: |-
  #     {
  # ...
  ```

* Query a lockup account's balance:

  ```bash
  laconicd query bank balances <address>

  # Example
  lockup_account_address=$(laconicd query auth module-account lps_lockup -o json | jq -r '.account.value.base_account.address')
  laconicd query bank balances $lockup_account_address
  balances:
  - amount: "1000"
    denom: alps
  pagination:
    total: "1"
  ```
