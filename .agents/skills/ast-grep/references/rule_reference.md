# ast-grep Rule Reference

This document provides comprehensive documentation for ast-grep rule syntax, covering all rule types and metavariables.

## Introduction to ast-grep Rules

ast-grep rules are declarative specifications for matching and filtering Abstract Syntax Tree (AST) nodes. They enable structural code search and analysis by defining conditions an AST node must meet to be matched.

### Rule Categories

ast-grep rules are categorized into three types:

* **Atomic Rules**: Match individual AST nodes based on intrinsic properties like code patterns (`pattern`), node type (`kind`), or text content (`regex`).
* **Relational Rules**: Define conditions based on a target node's position or relationship to other nodes (e.g., `inside`, `has`, `precedes`, `follows`).
* **Composite Rules**: Combine other rules using logical operations (AND, OR, NOT) to form complex matching criteria (e.g., `all`, `any`, `not`, `matches`).

## Anatomy of an ast-grep Rule Object

The ast-grep rule object is the core configuration unit defining how ast-grep identifies and filters AST nodes. It's typically written in YAML format.

### General Structure

Every field within an ast-grep Rule Object is optional, but at least one "positive" key (e.g., `kind`, `pattern`) must be present.

A node matches a rule if it satisfies all fields defined within that rule object, implying an implicit logical AND operation.

For rules using metavariables that depend on prior matching, explicit `all` composite rules are recommended to guarantee execution order.

### Rule Object Properties

| Property | Type | Category | Purpose | Example |
| :--- | :--- | :--- | :--- | :--- |
| `pattern` | String or Object | Atomic | Matches AST node by code pattern. | `pattern: console.log($ARG)` |
| `kind` | String | Atomic | Matches AST node by its kind name. | `kind: call_expression` |
| `regex` | String | Atomic | Matches node's text by Rust regex. | `regex: ^[a-z]+$` |
| `nthChild` | number, string, Object | Atomic | Matches nodes by their index within parent's children. | `nthChild: 1` |
| `range` | RangeObject | Atomic | Matches node by character-based start/end positions. | `range: { start: { line: 0, column: 0 }, end: { line: 0, column: 10 } }` |
| `inside` | Object | Relational | Target node must be inside node matching sub-rule. | `inside: { pattern: class $C { $$$ }, stopBy: end }` |
| `has` | Object | Relational | Target node must have descendant matching sub-rule. | `has: { pattern: await $EXPR, stopBy: end }` |
| `precedes` | Object | Relational | Target node must appear before node matching sub-rule. | `precedes: { pattern: return $VAL }` |
| `follows` | Object | Relational | Target node must appear after node matching sub-rule. | `follows: { pattern: import $M from '$P' }` |
| `all` | Array<Rule> | Composite | Matches if all sub-rules match. | `all: [ { kind: call_expression }, { pattern: foo($A) } ]` |
| `any` | Array<Rule> | Composite | Matches if any sub-rules match. | `any: [ { pattern: foo() }, { pattern: bar() } ]` |
| `not` | Object | Composite | Matches if sub-rule does not match. | `not: { pattern: console.log($ARG) }` |
| `matches` | String | Composite | Matches if predefined utility rule matches. | `matches: my-utility-rule-id` |

## Atomic Rules

Atomic rules match individual AST nodes based on their intrinsic properties.

### pattern: String and Object Forms

The `pattern` rule matches a single AST node based on a code pattern.

**String Pattern**: Directly matches using ast-grep's pattern syntax with metavariables.

```yaml
pattern: console.log($ARG)
```

**Object Pattern**: Offers granular control for ambiguous patterns or specific contexts.

* `selector`: Pinpoints a specific part of the parsed pattern to match.
  ```yaml
  pattern:
    selector: field_definition
    context: class { $F }
  ```

* `context`: Provides surrounding code context for correct parsing.

* `strictness`: Modifies the pattern's matching algorithm (`cst`, `smart`, `ast`, `relaxed`, `signature`).
  ```yaml
  pattern:
    context: foo($BAR)
    strictness: relaxed
  ```

### kind: Matching by Node Type

The `kind` rule matches an AST node by its `tree_sitter_node_kind` name, derived from the language's Tree-sitter grammar. Useful for targeting constructs like `call_expression` or `function_declaration`.

```yaml
kind: call_expression
```

### regex: Text-Based Node Matching

The `regex` rule matches the entire text content of an AST node using a Rust regular expression. It's not a "positive" rule, meaning it matches any node whose text satisfies the regex, regardless of its structural kind.

### nthChild: Positional Node Matching

The `nthChild` rule finds nodes by their 1-based index within their parent's children list, counting only named nodes by default.

* `number`: Matches the exact nth child. Example: `nthChild: 1`
* `string`: Matches positions using An+B formula. Example: `2n+1`
* `Object`: Provides granular control:
  * `position`: `number` or An+B string.
  * `reverse`: `true` to count from the end.
  * `ofRule`: An ast-grep rule to filter the sibling list before counting.

### range: Position-Based Node Matching

The `range` rule matches an AST node based on its character-based start and end positions. A `RangeObject` defines `start` and `end` fields, each with 0-based `line` and `column`. `start` is inclusive, `end` is exclusive.

## Relational Rules

Relational rules filter targets based on their position relative to other AST nodes. They can include `stopBy` and `field` options.

### inside: Matching Within a Parent Node

Requires the target node to be inside another node matching the `inside` sub-rule.

```yaml
inside:
  pattern: class $C { $$$ }
  stopBy: end
```

### has: Matching with a Descendant Node

Requires the target node to have a descendant node matching the `has` sub-rule.

```yaml
has:
  pattern: await $EXPR
  stopBy: end
```

### precedes and follows: Sequential Node Matching

* `precedes`: Target node must appear before a node matching the `precedes` sub-rule.
* `follows`: Target node must appear after a node matching the `follows` sub-rule.

Both include `stopBy` but not `field`.

### stopBy and field: Refining Relational Searches

**stopBy**: Controls search termination for relational rules.

* `"neighbor"` (default): Stops when immediate surrounding node doesn't match.
* `"end"`: Searches to the end of the direction (root for `inside`, leaf for `has`).
* `Rule object`: Stops when a surrounding node matches the provided rule (inclusive).

**field**: Specifies a sub-node within the target node that should match the relational rule. Only for `inside` and `has`.

**Best Practice**: When unsure, always use `stopBy: end` to ensure the search goes to the end of the direction.

## Composite Rules

Composite rules combine atomic and relational rules using logical operations.

### all: Conjunction (AND) of Rules

Matches a node only if all sub-rules in the list match. Guarantees order of rule matching, important for metavariables.

```yaml
all:
  - kind: call_expression
  - pattern: console.log($ARG)
```

### any: Disjunction (OR) of Rules

Matches a node if any sub-rules in the list match.

```yaml
any:
  - pattern: console.log($ARG)
  - pattern: console.warn($ARG)
  - pattern: console.error($ARG)
```

### not: Negation (NOT) of a Rule

Matches a node if the single sub-rule does not match.

```yaml
not:
  pattern: console.log($ARG)
```

### matches: Rule Reuse and Utility Rules

Takes a rule-id string, matching if the referenced utility rule matches. Enables rule reuse and recursive rules.

## Metavariables

Metavariables are placeholders in patterns to match dynamic content in the AST.

### $VAR: Single Named Node Capture

Captures a single named node in the AST.

* **Valid**: `$META`, `$META_VAR`, `$_`
* **Invalid**: `$invalid`, `$123`, `$KEBAB-CASE`
* **Example**: `console.log($GREETING)` matches `console.log('Hello World')`.
* **Reuse**: `$A == $A` matches `a == a` but not `a == b`.

### $$VAR: Single Unnamed Node Capture

Captures a single unnamed node (e.g., operators, punctuation).

**Example**: To match the operator in `a + b`, use `$$OP`.

```yaml
rule:
  kind: binary_expression
  has:
    field: operator
    pattern: $$OP
```

### $$$MULTI_META_VARIABLE: Multi-Node Capture

Matches zero or more AST nodes (non-greedy). Useful for variable numbers of arguments or statements.

* **Example**: `console.log($$$)` matches `console.log()`, `console.log('hello')`, and `console.log('debug:', key, value)`.
* **Example**: `function $FUNC($$$ARGS) { $$$ }` matches functions with varying parameters/statements.

### Non-Capturing Metavariables (_VAR)

Metavariables starting with an underscore (`_`) are not captured. They can match different content even if named identically, optimizing performance.

* **Example**: `$_FUNC($_FUNC)` matches `test(a)` and `testFunc(1 + 1)`.

### Important Considerations for Metavariable Detection

* **Syntax Matching**: Only exact metavariable syntax (e.g., `$A`, `$$B`, `$$$C`) is recognized.
* **Exclusive Content**: Metavariable text must be the only text within an AST node.
* **Non-working**: `obj.on$EVENT`, `"Hello $WORLD"`, `a $OP b`, `$jq`.

The ast-grep playground is useful for debugging patterns and visualizing metavariables.

## Common Patterns and Examples

### Finding Functions with Specific Content

Find functions that contain await expressions:

```yaml
rule:
  kind: function_declaration
  has:
    pattern: await $EXPR
    stopBy: end
```

### Finding Code Inside Specific Contexts

Find console.log calls inside class methods:

```yaml
rule:
  pattern: console.log($$$)
  inside:
    kind: method_definition
    stopBy: end
```

### Combining Multiple Conditions

Find async functions that use await but don't have try-catch:

```yaml
rule:
  all:
    - kind: function_declaration
    - has:
        pattern: await $EXPR
        stopBy: end
    - not:
        has:
          pattern: try { $$$ } catch ($E) { $$$ }
          stopBy: end
```

### Matching Multiple Alternatives

Find any type of console method call:

```yaml
rule:
  any:
    - pattern: console.log($$$)
    - pattern: console.warn($$$)
    - pattern: console.error($$$)
    - pattern: console.debug($$$)
```

## Troubleshooting Tips

1. **Rule doesn't match**: Use `dump_syntax_tree` to see the actual AST structure
2. **Relational rule issues**: Ensure `stopBy: end` is set for deep searches
3. **Wrong node kind**: Check the language's Tree-sitter grammar for correct kind names
4. **Metavariable not working**: Ensure it's the only content in its AST node
5. **Pattern too complex**: Break it down into simpler sub-rules using `all`
