# Household Cost Categories

Matches description line in a bank statement to a category, for budgeting
purposes. 

For example the text *New World Point Cheval Auckland Nz* should match to the
category *Food/Groceries*. So should *Woolworths Nz/Lynnmall New Lynn Nz*.

## Rules

Matching consists of user-supplied prefix rules. That is the rule matches if,
but only if, the rule matches exactly the start of the description.

For the examples above, the following rules would suffice:

    "New World"  => "Food/Groceries"
    "Woolworths" => "Food/Groceries"

Those rules would give the correct category for those supermarkets in other 
locations, say *Woolworths Nz/26 Custo Lwr Albert*
