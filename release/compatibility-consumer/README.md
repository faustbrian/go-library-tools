# Historical compatibility-set fixture

This is the frozen import-only fixture for published set
`golib-compat-v1-20260912.2`. Its original 123 direct module requirements,
Go minimum and 161 blank imports match the Tools v1.8.4 source and the
checksummed released catalogs. The appended indirect requirements and `go.sum`
were derived through authenticated public proxy/SumDB resolution during the
ownership cutover; they are not historical release assets.

Current selections and runtime composition tests now have one editable owner:
[go-library-assurance](https://github.com/faustbrian/go-library-assurance/tree/main/scenarios/compatibility-consumer).
The existing Tools generator and historical default path remain compatible.
Candidate/publication commands may deliberately update this fixture for a new
compatibility set; ordinary library releases need not update it.

Successful execution here proves compilation of the selected historical
imports, not current runtime/security qualification or a deployment
recommendation. Preserve historical set caveats and immutable publication
records. The active Assurance harness consumes current public releases without
being a library runtime dependency or release prerequisite.
