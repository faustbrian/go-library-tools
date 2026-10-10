// Package gates orchestrates repository checks from canonical module policy.
package gates

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/faustbrian/go-library-tools/v2/internal/config"
	"github.com/faustbrian/go-library-tools/v2/internal/coverage"
	"github.com/faustbrian/go-library-tools/v2/internal/docscheck"
	"github.com/faustbrian/go-library-tools/v2/internal/inventory"
	"github.com/faustbrian/go-library-tools/v2/internal/repositoryfile"
	"github.com/faustbrian/go-library-tools/v2/internal/services"
	"golang.org/x/mod/module"
)

const maximumMakefileSize = 4 << 20

const (
	maximumSecuritySourceFiles = 100_000
	maximumSecuritySourceSize  = 4 << 20
)

var (
	securityRuleList = regexp.MustCompile(`^G[0-9]{3}(?:\s*,\s*G[0-9]{3})*$`)
)

const (
	gitleaksPolicy = `title = "golib centrally owned secret scanning"
[extend]
useDefault = true

[[allowlists]]
description = "Exact independently regenerated public Webhook v1 interoperability records."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^testdata/vectors/v1\.json$''']
regexes = [
  '''^\n?      "key_mater(?:)ial_base64url": (?:)"Y3Jvc3MtbGFuZ3V(?:)hZ2UtdGVzdC1rZXk(?:)",$''',
  '''^\n?      "canonical(?:)_base64url": "d2(?:)ViaG9vay12MQphbG(?:)dvcml0aG06c2hhMj(?:)U2CnRpbWVzdGFtcD(?:)oxNzAwMDAwMDAwCm(?:)5vbmNlOmNIbDBhRz(?:)l1TFdacGVIUjFjbV(?:)V0Ym05dVkyVQprZX(?:)ktaWQ6YVc1MFpYSn(?:)ZjQzFyWlhrCm1ldG(?:)hvZDpwb3N0CnBhdG(?:)g6TDJodmIydHpMeV(?:)V5Um05eVpHVnljdw(?:)pxdWVyeTpZVDB5Sm(?:)1FOU1TWmlQWFIzYn(?:)labGJYQjBlVDAKaG(?:)9zdDpjbVZqWldsMl(?:)pYSXVaWGhoYlhCc1(?:)pUbzBORE0KY29udG(?:)VudC10eXBlOllYQn(?:)diR2xqWVhScGIyNH(?:)Zhbk52YmcKaWRlbX(?:)BvdGVuY3kta2V5Ol(?:)pYWmxiblF0YVc1MF(?:)pYSnZjQzB4CmJvZH(?:)ktc2hhMjU2OmFrZk(?:)RHM3Q4TzVvZHZKWU(?:)dhZlJuVE9DSXlQeW(?:)Rtazktbjh3X2FvSD(?:)N1R3cKbWV0YWRhdG(?:)E6WkVkV2RWbFhOVE(?:)E5V1ZkT2RGcFJDbV(?:)JYTlhCWk1qbHJXbE(?:)U5WXpJMWRtUXlNV2(?:)hpYVVScGJVbE4K",$''',
  '''^\n?      "canonical(?:)_base64url": "d2(?:)ViaG9vay12MQphbG(?:)dvcml0aG06c2hhNT(?:)EyCnRpbWVzdGFtcD(?:)oxNzAwMDAwMDAwCm(?:)5vbmNlOmNIbDBhRz(?:)l1TFdacGVIUjFjbV(?:)V0Ym05dVkyVQprZX(?:)ktaWQ6YVc1MFpYSn(?:)ZjQzFyWlhrCm1ldG(?:)hvZDpwb3N0CnBhdG(?:)g6TDJodmIydHpMeV(?:)V5Um05eVpHVnljdw(?:)pxdWVyeTpZVDB5Sm(?:)1FOU1TWmlQWFIzYn(?:)labGJYQjBlVDAKaG(?:)9zdDpjbVZqWldsMl(?:)pYSXVaWGhoYlhCc1(?:)pUbzBORE0KY29udG(?:)VudC10eXBlOllYQn(?:)diR2xqWVhScGIyNH(?:)Zhbk52YmcKaWRlbX(?:)BvdGVuY3kta2V5Ol(?:)pYWmxiblF0YVc1MF(?:)pYSnZjQzB4CmJvZH(?:)ktc2hhMjU2OmFrZk(?:)RHM3Q4TzVvZHZKWU(?:)dhZlJuVE9DSXlQeW(?:)Rtazktbjh3X2FvSD(?:)N1R3cKbWV0YWRhdG(?:)E6WkVkV2RWbFhOVE(?:)E5V1ZkT2RGcFJDbV(?:)JYTlhCWk1qbHJXbE(?:)U5WXpJMWRtUXlNV2(?:)hpYVVScGJVbE4K",$''',
  '''^\n?      "canonical_base64url": "webhook-v1\nalgorithm:sha256\ntimestamp:1700000000\nnonce:cHl0aG9uLWZpeHR1cmUtbm9uY2U\nkey-id:aW50ZXJvcC1rZXk\nmethod:post\npath:L2hvb2tzLyUyRm9yZGVycw\nquery:YT0yJmE9MSZiPXR3byZlbXB0eT0\nhost:cmVjZWl2ZXIuZXhhbXBsZTo0NDM\ncontent-type:YXBwbGljYXRpb24vanNvbg\nidempotency-key:ZXZlbnQtaW50ZXJvcC0x\nbody-sha256:akfDG3t8O5odvJYGafRnTOCIyPydmk9-n8w_aoH3uGw\nmetadata:ZEdWdVlXNTA9WVdOdFpRCmRXNXBZMjlrWlE9YzI1dmQyMWhiaURpbUlN\n",$''',
  '''^\n?      "canonical_base64url": "webhook-v1\nalgorithm:sha512\ntimestamp:1700000000\nnonce:cHl0aG9uLWZpeHR1cmUtbm9uY2U\nkey-id:aW50ZXJvcC1rZXk\nmethod:post\npath:L2hvb2tzLyUyRm9yZGVycw\nquery:YT0yJmE9MSZiPXR3byZlbXB0eT0\nhost:cmVjZWl2ZXIuZXhhbXBsZTo0NDM\ncontent-type:YXBwbGljYXRpb24vanNvbg\nidempotency-key:ZXZlbnQtaW50ZXJvcC0x\nbody-sha256:akfDG3t8O5odvJYGafRnTOCIyPydmk9-n8w_aoH3uGw\nmetadata:ZEdWdVlXNTA9WVdOdFpRCmRXNXBZMjlrWlE9YzI1dmQyMWhiaURpbUlN\n",$''',
]

[[allowlists]]
description = "Exact Verkle source checksums recomputed from the pinned public upstream revision."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^specification/sources\.json$''']
regexes = [
  '''^\n?        "api\.go"(?:): "dac8a950e16df(?:)2bed91979615b917(?:)e236b701930b0a76(?:)44fd72ebf0c38d46(?:)7b0",$''',
  '''^\n?        "api_eip(?:)7594\.go": "ba5f2(?:)a0739cb66f87f095(?:)ceb150fab1b6e4de(?:)1cbb3c32200ebd18(?:)a781fe2360d",$''',
]


[[allowlists]]
description = "Exact public Password decision checksum records independently recomputed from immutable decision registers."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^CHANGELOG\.md$''']
regexes = [
  '''^\n?  PASSWORD-DEC-0(?:)06 sha256:2aa164(?:)80aefc694ed42f78(?:)109e6347c3fd6610(?:)5fee868362aef58c(?:)087368bf83;$''',
  '''^\n?  PASSWORD-DEC-0(?:)07 sha256:23b884(?:)6b5751879f08ba17(?:)02f0fa3625b5d5cb(?:)b6e9d9fe46536f4b(?:)bb25270d42\.$''',
  '''^\n?  PASSWORD-DEC-0(?:)01 sha256:81d693(?:)f05f84c603efdb62(?:)9d9249e13e134def(?:)4e645afffd12e50b(?:)1e7233823c;$''',
  '''^\n?  PASSWORD-DEC-0(?:)02 sha256:d0aff4(?:)5bb407811912bc21(?:)be738605732adbd6(?:)04decc339b449bd2(?:)5fdf125ddf;$''',
  '''^\n?  PASSWORD-DEC-0(?:)03 sha256:abb7d2(?:)b09fe08461078fbe(?:)be88637d317ef52b(?:)5ffc035fa3072950(?:)717978f24f;$''',
  '''^\n?  PASSWORD-DEC-0(?:)04 sha256:35d8b7(?:)5df3f64f7a8d45e3(?:)8c763ebc7e110953(?:)b0542d3c03cd230c(?:)2d1194e83e;$''',
  '''^\n?  PASSWORD-DEC-0(?:)05 sha256:ed3cb3(?:)293b84a79a7cd068(?:)92599874eaf56b4f(?:)91071e8537df71fc(?:)93794a7104;$''',
  '''^\n?  PASSWORD-DEC-0(?:)06 sha256:d02071(?:)7a693732de00655a(?:)0a0c4359626feacc(?:)6f71d130acbd2180(?:)0e3979a748;$''',
  '''^\n?  PASSWORD-DEC-0(?:)07 sha256:e3dcde(?:)e13b077e6effd51a(?:)4b6dd9d00c0b3a37(?:)3826aaa0ee305495(?:)e4088a5def\.$''',
  '''^\n?  PASSWORD-DEC-0(?:)01 sha256:ba44cc(?:)58be46506940728e(?:)9d6edfcfc165f84c(?:)201ac01efe709b49(?:)49a211cd92;$''',
  '''^\n?  PASSWORD-DEC-0(?:)02 sha256:9d9202(?:)7d805ca587dcce98(?:)75bb0af09d2d9913(?:)ad883bee56e17340(?:)7d7a6b6362;$''',
  '''^\n?  PASSWORD-DEC-0(?:)03 sha256:ea20f1(?:)c77cbbb8f8b5ca43(?:)a23c217ae77e142e(?:)ed42ee4db1f636be(?:)1fc949ab45;$''',
  '''^\n?  PASSWORD-DEC-0(?:)04 sha256:3526e9(?:)549c855d048fec5c(?:)31220d1c3085331e(?:)ef128437c1d3bc1a(?:)1e682daf9d;$''',
  '''^\n?  PASSWORD-DEC-0(?:)05 sha256:695d18(?:)94768e03f3a3d599(?:)4e432d9eac0248b9(?:)81bc3493c0931ce9(?:)ba19996d84;$''',
]

[[allowlists]]
description = "Exact historical Password APIDIFF tool version is a public module identity."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^tools/versions\.env$''']
regexes = [
  '''^\n?APIDIFF_VERSION=(?:)v0\.0\.0-202607182(?:)01538-764159d718(?:)ef$''',
]

[[allowlists]]
description = "Exact temp-clone-token examples in the pinned public GitHub OpenAPI artifact."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^specification/independent/github-rest-api/api\.github\.com\.2022-11-28\.json$''']
regexes = [
  '''^\n?              "t(?:)emp_clone_token"(?:): "ABTLWHOULUVAX(?:)GTRYU7OC2876QJ2O(?:)",$''',
  '''^\n?                (?:)"temp_clone_toke(?:)n": "ABTLWHOULUV(?:)AXGTRYU7OC2876QJ(?:)2O",$''',
  '''^\n?            "tem(?:)p_clone_token": (?:)"ABTLWHOULUVAXGT(?:)RYU7OC2876QJ2O",$''',
  '''^\n?          "temp_(?:)clone_token": "A(?:)BTLWHOULUVAXGTRY(?:)U7OC2876QJ2O",$''',
]

[[allowlists]]
description = "The immutable CI tooling checkout is verified separately."
paths = ['''^\.golib-tooling(?:/|$)''']

[[allowlists]]
description = "Exact public API Query v1 baseline assertion, independently recomputed from the released Git blob."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^\.verification/cohesion/api-query-v1-oracle/verify\.go$''']
regexes = ['''^\n?	validateRelease(?:)Artifact\(ctx, ro(?:)ot, releaseCommi(?:)t, packet\.Releas(?:)e\.APIBaseline, "(?:)ec28983aeaa00eed(?:)bfa61d2a6c5eeb67(?:)fc6c3185", 25021(?:), "b8eac53159059(?:)9de50da2ed042591(?:)db7edf07c6a438e9(?:)00085aca53ed6354(?:)fc5"\)$''']

[[allowlists]]
description = "Exact deterministic public Capability interoperability token signed with the published test key."
condition = "AND"
targetRules = ["jwt"]
regexTarget = "line"
paths = ['''^testdata/v1-hmac\.token$''']
regexes = ['''^\n?cap1\.eyJ2IjoxLCJ(?:)0eXAiOiJjYXBhYml(?:)saXR5IiwiYWxnIjo(?:)iaG1hYy1zaGEyNTY(?:)iLCJraWQiOiJpbnR(?:)lcm9wIn0\.eyJ2Ijo(?:)xLCJpc3MiOiJpbnR(?:)lcm9wIiwiYXVkIjp(?:)bInNlcnZpY2UiXSw(?:)iYmVhcmVyIjp0cnV(?:)lLCJyZXNvdXJjZSI(?:)6Im9iamVjdHMvNDI(?:)iLCJvcGVyYXRpb24(?:)iOiJyZWFkIiwiaWF(?:)0IjoxNzg2Mjc2ODA(?:)wLCJuYmYiOjE3ODY(?:)yNzY4MDAsImV4cCI(?:)6MTc4NjI3Njg2MCw(?:)iaWQiOiJpbnRlcm9(?:)wLWNhcGFiaWxpdHk(?:)ifQ\.Iwj0h9OnGkC0(?:)P1Hrw2L9dweW8Wvs(?:)C0md1z65DmNlXTQ$''']

[[allowlists]]
description = "Pinned apidiff versions in exact compatibility rehearsal Makefiles are tool identities."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "secret"
regexes = ['''^v0\.0\.0-[0-9]{14}-[0-9a-f]{12}$''']
paths = [
  '''^internal/gates/api\.go$''',
  '''^rehearsals/go-(?:authorization|openapi)/verification/package\.mk$''',
]

[[allowlists]]
description = "Historical APIDIFF_VERSION pseudo-version in the retired legacy tool-version file."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "secret"
regexes = ['''^v0\.0\.0-[0-9]{14}-[0-9a-f]{12}$''']
paths = ['''^\.golib/versions\.env$''']

[[allowlists]]
description = "Historical APIDIFF_VERSION assignments in exact retired Makefiles are public tool identities."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
regexes = ['''^\n?APIDIFF_VERSION[ \t]*[?:]=[ \t]*v0\.0\.0-[0-9]{14}-[0-9a-f]{12}[ \t]*$''']
paths = ['''^(?:\.golib/package\.mk|verification/package\.mk|Makefile)$''']

[[allowlists]]
description = "Exact synthetic Stripe token used by hostile inventory identity tests."
condition = "AND"
targetRules = ["stripe-access-token"]
regexTarget = "secret"
regexes = ['''^sk_test_0123456789abcdefghijklmnopqrstuv$''']
paths = ['''^internal/inventory/inventory_test\.go$''']

[[allowlists]]
description = "Exact public Ethereum legacy trie fixture line."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^testdata/ethereum-tests/TrieTests/trietest\.json$''']
regexes = ['''^\n?      \[ "key1", "01234567(?:)89012345(?:)67890123(?:)45678901(?:)23456789(?:)Very_Lon(?:)g"\],$''']

[[allowlists]]
description = "Exact public signing fixture in the shared HTTP-signature differential corpus."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "secret"
regexes = ['''^01234567(?:)89abcdef(?:)01234567(?:)89abcdef(?:)01234567(?:)89abcdef(?:)01234567(?:)89abcdef$''']
paths = ['''^differential/shared-corpus/corpus_test\.go$''']

[[allowlists]]
description = "Exact immutable Confluent decision digests recorded in the provider changelog."
condition = "AND"
targetRules = ["confluent-secret-key"]
regexTarget = "secret"
regexes = [
  '''^67b4c198(?:)5e70a7ae(?:)a45d754e(?:)14be8468(?:)83a37ccd(?:)073d76e5(?:)99d71900(?:)4af0ea37$''',
  '''^4c9ab0b7(?:)2db6bcd6(?:)a6f90cd8(?:)e638e7f2(?:)80708c95(?:)18128b59(?:)33a9a904(?:)ad072ff7$''',
  '''^c92530ae(?:)87091474(?:)8c82e238(?:)b72ff1d3(?:)c60c2ec0(?:)360150b0(?:)1863750e(?:)6dad0ac1$''',
]
paths = ['''^providers/confluent/CHANGELOG\.md$''']

[[allowlists]]
description = "Exact external-sort threat-model resource-policy prose is not a credential."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^docs/threat-model\.md$''']
regexes = ['''^\n?parent directory, AES-256 key, record/(?:)chunk/(?:)population limits, context, and$''']

[[allowlists]]
description = "Exact public Authentication decision ID and SHA-256 records in both immutable decision histories."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^CHANGELOG\.md$''']
regexes = [
  '''^\n?- AUTH-DEC-011 sha256:be33f511(?:)1184dec4(?:)c31bdeac(?:)3760ee2e(?:)f0fcdd73(?:)0a677f55(?:)479e7ee8(?:)f66ad659$''',
  '''^\n?- AUTH-DEC-012 sha256:91356766(?:)40d9c0df(?:)84a1b676(?:)985d9f5c(?:)25129983(?:)04591b88(?:)09092b50(?:)ce0d7884$''',
  '''^\n?- AUTH-DEC-001 sha256:dbf293fd(?:)525d952d(?:)fe51ac1a(?:)f7a43325(?:)429fa00d(?:)fe6d34e2(?:)cdc45069(?:)830330f5$''',
  '''^\n?- AUTH-DEC-002 sha256:5d9b14d2(?:)4d6ac9ea(?:)1f84b43d(?:)5d745454(?:)b53e5271(?:)fee035aa(?:)909d1d9a(?:)2342d5fe$''',
  '''^\n?- AUTH-DEC-002 sha256:15e8ade4(?:)98da8125(?:)48bf36de(?:)b3ab0b77(?:)d5fae914(?:)62036bc3(?:)6e27aaa3(?:)9fb4f273$''',
  '''^\n?- AUTH-DEC-003 sha256:81ffead7(?:)7cfa5491(?:)3477444f(?:)24798f9b(?:)632beb94(?:)057e578e(?:)c1f266b0(?:)20c6a7c1$''',
  '''^\n?- AUTH-DEC-003 sha256:10e348b7(?:)b49b28fb(?:)53f35f00(?:)72cbda5f(?:)f08156f8(?:)71c854fc(?:)b7e0f62b(?:)70f0ed9e$''',
  '''^\n?- AUTH-DEC-004 sha256:86762f7d(?:)564203e9(?:)714e51b8(?:)67aa8ec4(?:)ebc913d0(?:)ca704e71(?:)6a409e56(?:)b2537722$''',
  '''^\n?- AUTH-DEC-004 sha256:06998d20(?:)abf14867(?:)af6c7195(?:)a88b8ada(?:)dc4a66dd(?:)a00e1990(?:)172180c5(?:)c71a61a6$''',
  '''^\n?- AUTH-DEC-005 sha256:b8d53a0e(?:)3c098089(?:)2e5b161e(?:)2e5b2534(?:)f635d250(?:)e4310757(?:)d11d78ff(?:)08b773c4$''',
  '''^\n?- AUTH-DEC-006 sha256:e18d6391(?:)38c2828e(?:)3b8240a6(?:)bf80c543(?:)7b04d630(?:)da12cc3e(?:)419d476a(?:)b35b88e7$''',
  '''^\n?- AUTH-DEC-007 sha256:f8b3f8e8(?:)393811a2(?:)c3ce0698(?:)8b92f812(?:)52babeed(?:)e518c683(?:)6afaac7a(?:)3de89326$''',
  '''^\n?- AUTH-DEC-007 sha256:e9a92111(?:)6a9efa03(?:)3845aff4(?:)eb3e9313(?:)03ad4f13(?:)80b19033(?:)8c730245(?:)ea8860d6$''',
  '''^\n?- AUTH-DEC-008 sha256:b4f02bec(?:)b2f6f0f2(?:)d098c8d2(?:)4fe75004(?:)8bf87f8d(?:)2aefc8c6(?:)80745c0d(?:)240c812a$''',
  '''^\n?- AUTH-DEC-009 sha256:0104da4c(?:)0dca68fb(?:)7146d17f(?:)5d70077f(?:)e34edaf9(?:)19feef29(?:)b718252e(?:)db65ee37$''',
  '''^\n?- AUTH-DEC-010 sha256:4a94dc00(?:)885603ca(?:)190f6e2e(?:)da39a98d(?:)09511ff7(?:)417bb3de(?:)9cb13177(?:)4398d60f$''',
]

[[allowlists]]
description = "Exact public Authentication JWT decision table row naming authority and executable tests."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^jwt/specification/README\.md$''']
regexes = ['''^\n?\| \[JWT-DEC-004\]\(\.\./docs/specification-decisions\.md\) \| rfc7518-source, rfc7517-source, rfc8725-source \| TestValidatorRejectsCryptographicallyUnsafeKeys, TestValidateKeyMaterialRejectsEveryInvalidRepresentation, TestRemoteJWKValidationRejectsEveryKey(?:)PolicyViolation, TestRFC(?:)7520HMACJWKInteroperability \|$''']

[[allowlists]]
description = "Exact public JSONAPI specification decision ID and SHA-256 records, independently recomputed from immutable decision registers."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^CHANGELOG\.md$''']
regexes = [
  '''^\n?- JSONAPI-DEC-011 sha256:d234cd9a(?:)4e40ce09(?:)214a023d(?:)09c11af6(?:)e7077c1b(?:)a5a39f8c(?:)01a9dc6c(?:)bbb98095$''',
  '''^\n?- JSONAPI-DEC-012 sha256:85161bb4(?:)7428b3ae(?:)ad1e8930(?:)b318ccd9(?:)a8146800(?:)1e190378(?:)cdb7b621(?:)eb3022ef$''',
  '''^\n?- JSONAPI-DEC-013 sha256:9b156b1a(?:)af44e12e(?:)b6f10ee7(?:)4a737c98(?:)1aec3f89(?:)8b94fad1(?:)f547ac2e(?:)68c57e04$''',
  '''^\n?- JSONAPI-DEC-001 sha256:7ab9e941(?:)c1eed52a(?:)bff8d37e(?:)d674e466(?:)d75682bf(?:)c9dc3427(?:)24208969(?:)2c7fadd4$''',
  '''^\n?- JSONAPI-DEC-002 sha256:92b24e46(?:)b91b1722(?:)39d6c7ef(?:)6afeab30(?:)28f6d3d8(?:)83f6c66e(?:)641b9515(?:)74811cb3$''',
  '''^\n?- JSONAPI-DEC-003 sha256:b894d768(?:)df3cdfa9(?:)6914f4b8(?:)805c3dfd(?:)fe3bf398(?:)8b7281e4(?:)8f73113c(?:)7abff5c1$''',
  '''^\n?- JSONAPI-DEC-004 sha256:eefa54a5(?:)8ed63f71(?:)bbaa3b6d(?:)2d8be4ca(?:)bcaef92e(?:)e05d1d9c(?:)7e7014ae(?:)bde062fd$''',
  '''^\n?- JSONAPI-DEC-005 sha256:8c5339c0(?:)c40d1ca9(?:)ecbc7db8(?:)fe8ecb13(?:)3d597b6d(?:)fa4d535f(?:)9140e7a8(?:)03751050$''',
  '''^\n?- JSONAPI-DEC-006 sha256:91709eb6(?:)d8259986(?:)3d8e0375(?:)9d1cd9f7(?:)268d26e8(?:)cfb3b0dd(?:)652b7e27(?:)1873cb97$''',
  '''^\n?- JSONAPI-DEC-007 sha256:17f841ab(?:)f7b95111(?:)69d91641(?:)2f03e7b5(?:)13bfde2b(?:)93b76422(?:)66a23d92(?:)e5ee99f0$''',
  '''^\n?- JSONAPI-DEC-008 sha256:383a65c3(?:)c89cae68(?:)d2d61499(?:)3caab29d(?:)a026c9a6(?:)a2c057e5(?:)70b6bbf6(?:)52c9a7db$''',
  '''^\n?- JSONAPI-DEC-009 sha256:ee353253(?:)3128439a(?:)7617e7da(?:)04fbbda1(?:)3f30b14b(?:)7eb30119(?:)0fd5712a(?:)1d2c0625$''',
  '''^\n?- JSONAPI-DEC-010 sha256:6721f9d3(?:)ea09062d(?:)9665828d(?:)3e1f2127(?:)d3bd3d1d(?:)967d6214(?:)e01a164b(?:)c7b8b1f2$''',
  '''^\n?- JSONAPI-DEC-001 sha256:1c992612(?:)f6fdf57e(?:)58587537(?:)47bcb574(?:)64f149b5(?:)17084813(?:)681e7029(?:)f31186f0$''',
  '''^\n?- JSONAPI-DEC-002 sha256:7d6658ae(?:)3e8b8176(?:)dffdb809(?:)96a5439f(?:)55b30d8d(?:)3c1e5a2a(?:)3d1c9f28(?:)5f2d2338$''',
  '''^\n?- JSONAPI-DEC-003 sha256:86ae12b7(?:)a1ba561c(?:)f2f69475(?:)6d980ed6(?:)eec544b3(?:)007f4705(?:)8e87f8c5(?:)1023115e$''',
  '''^\n?- JSONAPI-DEC-004 sha256:be285995(?:)51762934(?:)5b407e0b(?:)93744ada(?:)5cc85d54(?:)b72092b3(?:)8310d9ff(?:)16b795f5$''',
  '''^\n?- JSONAPI-DEC-005 sha256:db4077bb(?:)84f3aadc(?:)3f3b68c3(?:)07029f01(?:)c2f49ae4(?:)5a8f17d3(?:)4e0d4945(?:)a03f31d3$''',
  '''^\n?- JSONAPI-DEC-006 sha256:526de3bc(?:)95e292ca(?:)970e560b(?:)d06783a9(?:)ce70a8ef(?:)656dfd8e(?:)7ef51151(?:)a00339c5$''',
  '''^\n?- JSONAPI-DEC-007 sha256:ad587c6d(?:)3623a796(?:)4b3047f7(?:)42d2b9c1(?:)91b9b979(?:)98417185(?:)7c727709(?:)2d110501$''',
  '''^\n?- JSONAPI-DEC-008 sha256:98b181e4(?:)0f339a4a(?:)11b389a6(?:)bfe2ed0c(?:)a28ac122(?:)5fabfdaa(?:)d62f5430(?:)bb5363b7$''',
  '''^\n?- JSONAPI-DEC-009 sha256:47736a06(?:)49bf48e1(?:)9fed8cc7(?:)5bb8308a(?:)c1790699(?:)487a4e40(?:)55a8e2a4(?:)d486ec2b$''',
  '''^\n?- JSONAPI-DEC-010 sha256:c7a9604e(?:)5ef3e66a(?:)7576826b(?:)8ab08b6e(?:)58d5de1b(?:)d4d95451(?:)84b4ff0c(?:)66d8c2fa$''',
]
[[allowlists]]
description = "Exact public OpenAPI decision ID and SHA-256 records, independently recomputed from immutable decision registers."
condition = "AND"
targetRules = ["generic-api-key"]
regexTarget = "line"
paths = ['''^CHANGELOG\.md$''']
regexes = [
  '''^\n?- OPENAPI-DEC-001 sha256:945adbbd(?:)c8784d07(?:)8c0c92f1(?:)8e25f093(?:)35b55c60(?:)f4dabc6c(?:)4cd892c4(?:)edf8e227$''',
  '''^\n?- OPENAPI-DEC-002 sha256:7a6781bc(?:)ac02f255(?:)55164496(?:)9f0b5331(?:)0bddcb9c(?:)417eaaf7(?:)062ecb0e(?:)b421a193$''',
  '''^\n?- OPENAPI-DEC-003 sha256:13253357(?:)9e14d96e(?:)ccffbf6c(?:)9456d1a9(?:)e36d895e(?:)ffb2b977(?:)d4deffdf(?:)a7e11e22$''',
  '''^\n?- OPENAPI-DEC-004 sha256:3412d17f(?:)04d660ca(?:)df876062(?:)ff3f6775(?:)31a6e928(?:)d81ebf6e(?:)98926f74(?:)ef9a5525$''',
  '''^\n?- OPENAPI-DEC-005 sha256:457129d9(?:)ab2d52e9(?:)0b4f2dff(?:)335b8cc3(?:)d0faeee8(?:)dc104f0e(?:)78820ab8(?:)b7d0b2da$''',
  '''^\n?- OPENAPI-DEC-006 sha256:e0f1f7cd(?:)1e4806a3(?:)0399749f(?:)079f3700(?:)530c70f8(?:)ef8bf86e(?:)6ef92924(?:)57154b70$''',
  '''^\n?- OPENAPI-DEC-007 sha256:514cb998(?:)9d61158b(?:)fb863b95(?:)37ae21a2(?:)f8a8682c(?:)e5baf327(?:)4b220ab5(?:)00a9fbed$''',
  '''^\n?- OPENAPI-DEC-008 sha256:cc10809f(?:)406b1233(?:)6deb5353(?:)7ad9146b(?:)a71ac98c(?:)6962779c(?:)229dc203(?:)c501e018$''',
  '''^\n?- OPENAPI-DEC-009 sha256:962eee67(?:)6c2080c6(?:)558e03d5(?:)3fe08911(?:)ab6cc2e9(?:)de353428(?:)f7f4cef6(?:)659bc300$''',
  '''^\n?- OPENAPI-DEC-010 sha256:192a09d2(?:)7fcf47ca(?:)576b13c5(?:)37c30266(?:)880ca429(?:)8db996e4(?:)00c41f3f(?:)fbec96af$''',
  '''^\n?- OPENAPI-DEC-001 sha256:5c40c0ec(?:)0ffe030e(?:)87f09461(?:)57b1670d(?:)d6e681ca(?:)fbd45833(?:)86df6698(?:)94d39882$''',
  '''^\n?- OPENAPI-DEC-004 sha256:c38e2c7b(?:)778c8636(?:)f4ed8c9b(?:)7231a104(?:)9cfac298(?:)59a655b5(?:)c7011d44(?:)159bd1e9$''',
  '''^\n?- OPENAPI-DEC-006 sha256:df83163c(?:)b3d52899(?:)abcdfcfc(?:)d95245e3(?:)ea9914b4(?:)1401ef55(?:)2eb6dc88(?:)12ded7bb$''',
  '''^\n?- OPENAPI-DEC-010 sha256:cfd3ed8d(?:)4b897ba3(?:)9715ecd0(?:)61f875b6(?:)d5da62eb(?:)87aab1a0(?:)17de8e9f(?:)79fb37a7$''',
]

`
	analysisSecurityPolicy = `version: 1
rules:
  security/no-unsafe:
    status: blocking
    promotion:
      version: 1.0.0
      evidence: ecosystem security policy prohibits unsafe, cgo, and go:linkname bypasses
`
)

const (
	golangCILintVersion = "v2.14.0"
	staticcheckVersion  = "v0.8.1"
	nilAwayVersion      = "v0.0.0-20260720194628-9fd1b8d7bac8"
	govulncheckVersion  = "v1.6.0"
	gosecVersion        = "v2.29.0"
	goAnalysisVersion   = "v1.0.0"
	gitleaksVersion     = "v8.30.1"
	goLicensesVersion   = "v2.0.1"
	cycloneDXVersion    = "v1.10.0"
)

// Command is one external process invocation without shell interpretation.
type Command struct {
	// boundedScanner is only set by owned pinned-scanner/source-acquisition
	// callers. It does not claim containment of arbitrary repository code.
	boundedScanner bool
	Name           string
	Args           []string
	Dir            string
	Env            map[string]string
	Stdin          io.Reader
	Stdout         io.Writer
	Stderr         io.Writer
}

const maximumSecurityProcessOutput = 4 << 20

var errRepositoryGitleaksIgnore = errors.New("security-policy: repository-owned .gitleaksignore is not permitted for security-enabled checks")

type boundedProcessOutput struct {
	mutex    sync.Mutex
	limit    int
	written  int
	overflow bool
	onLimit  func()
	// Only a fixed owned terminal line may be recognized; stream text is never retained.
	terminalLine       string
	terminalLineOffset int
	terminalLineMatch  bool
	metadata           *secretMetadata
}

func (output *boundedProcessOutput) Write(value []byte) (int, error) {
	output.mutex.Lock()
	if len(value) > output.limit-output.written {
		output.written = output.limit
		first := !output.overflow
		output.overflow = true
		onLimit := output.onLimit
		output.mutex.Unlock()
		if first && onLimit != nil {
			onLimit()
		}
		return len(value), nil
	}
	output.written += len(value)
	if output.metadata != nil {
		output.metadata.write(value)
	}
	if output.terminalLine != "" {
		for _, character := range value {
			if character == '\n' {
				output.terminalLineMatch = output.terminalLineOffset == len(output.terminalLine)
				output.terminalLineOffset = 0
				continue
			}
			output.terminalLineMatch = false
			if output.terminalLineOffset >= 0 && output.terminalLineOffset < len(output.terminalLine) && character == output.terminalLine[output.terminalLineOffset] {
				output.terminalLineOffset++
			} else {
				output.terminalLineOffset = -1
			}
		}
	}
	output.mutex.Unlock()
	return len(value), nil
}

func (output *boundedProcessOutput) setOverflowCallback(callback func()) {
	output.mutex.Lock()
	output.onLimit = callback
	alreadyOverflowed := output.overflow
	output.mutex.Unlock()
	if alreadyOverflowed && callback != nil {
		callback()
	}
}

func (output *boundedProcessOutput) didOverflow() bool {
	output.mutex.Lock()
	defer output.mutex.Unlock()
	return output.overflow
}

func (output *boundedProcessOutput) matchedTerminalLine() bool {
	output.mutex.Lock()
	defer output.mutex.Unlock()
	return !output.overflow && output.terminalLineMatch
}

func (output *boundedProcessOutput) findingMetadata() string {
	output.mutex.Lock()
	defer output.mutex.Unlock()
	if output.overflow || output.metadata == nil {
		return ""
	}
	return output.metadata.summary()
}

// Executor runs one external command.
type Executor interface {
	Run(context.Context, Command) error
}

type taskWorkspace interface {
	TemporaryDirectory() string
}

// Runner executes gates for modules in one validated repository.
type Runner struct {
	Root              string
	Catalog           inventory.Inventory
	Policy            config.Config
	Executor          Executor
	Output            io.Writer
	coverageFiles     coverageFileSystem
	secretConfigFiles secretConfigFileSystem
	apiFiles          apiFileSystem
	apiReadBaseline   func(string, string, int64) ([]byte, error)
	releaseFiles      releaseFileSystem
	releaseArchive    func(io.Writer, module.Version, string) error
	mutationFiles     mutationFileSystem
	mutationCampaign  mutationCampaignRunner
	mutationImport    mutationImportRunner
	startServices     serviceStarter
	serviceHTTPProbe  services.HTTPProbe
	serviceIdentities map[string]string
	// DocumentationSpelling is an isolated test boundary. Production callers
	// leave it nil and use the pinned task-owned implementation.
	DocumentationSpelling func(context.Context, string) error
	// DocumentationLinks is an isolated test boundary. Production callers
	// leave it nil and use the checksum-pinned task-owned implementation.
	DocumentationLinks       func(context.Context, string) error
	documentationRelease     func(string, string) (docscheck.LycheeRelease, error)
	documentationExtract     func(string, docscheck.LycheeRelease) ([]byte, error)
	gitleaksSourceCleanup    func(string) error
	securitySourceLimits     *securitySourceLimits
	repositorySecretsScanned bool
}

type namedWriteCloser interface {
	io.WriteCloser
	Name() string
}

type coverageFileSystem interface {
	CreateTemp(string) (namedWriteCloser, error)
	Open(string) (io.ReadCloser, error)
	Remove(string) error
}

type operatingCoverageFiles struct{}

func (operatingCoverageFiles) CreateTemp(directory string) (namedWriteCloser, error) {
	return os.CreateTemp(directory, "golib-coverage-*.out")
}

func (operatingCoverageFiles) Open(path string) (io.ReadCloser, error) {
	// #nosec G304 -- callers open the exact task-owned coverage profile path created by this gate
	return os.Open(path)
}

func (operatingCoverageFiles) Remove(path string) error {
	return os.Remove(path)
}

type secretConfigFileSystem interface {
	CreateTemp(string, string) (namedWriteCloser, error)
	Remove(string) error
}

type operatingSecretConfigFiles struct{}

func (operatingSecretConfigFiles) CreateTemp(directory, pattern string) (namedWriteCloser, error) {
	return os.CreateTemp(directory, pattern)
}

func (operatingSecretConfigFiles) Remove(path string) error {
	return os.Remove(path)
}

// Secrets scans repository history and the current tree with the same bounded,
// sanitized policies used by the standard contract, without runtime gates.
func (runner Runner) Secrets(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	return runner.runRepositorySecrets(ctx, output, ".")
}

// Check runs the standard contract for each explicitly selected module.
func (runner Runner) Check(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	if err := runner.preflightSecurityPolicies(ctx, output, modules); err != nil {
		return err
	}
	if err := runner.scanSelectedSecurity(ctx, output, modules); err != nil {
		return err
	}
	return runner.checkModules(ctx, output, modules)
}

func (runner Runner) checkModules(ctx context.Context, output io.Writer, modules []inventory.Module) error {
	for _, module := range modules {
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			return scoped.checkModule(ctx, output, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

// Local runs the bounded repository-owned checks used for ordinary pull
// requests. Expensive evidence gates remain explicit Check operations selected
// for a material risk or release milestone.
func (runner Runner) Local(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	if err := runner.preflightSecurityPolicies(ctx, output, modules); err != nil {
		return err
	}
	if err := runner.scanSelectedSecurity(ctx, output, modules); err != nil {
		return err
	}
	for _, module := range modules {
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			return scoped.checkModuleLocal(ctx, output, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

// Coverage collects production-package coverage using each module's acceptance
// policy. Omission preserves exact enforcement.
func (runner Runner) Coverage(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["coverage"] {
			_, _ = fmt.Fprintf(output, "[%s] coverage: not applicable\n", module.Directory)
			continue
		}
		if err := runner.withModuleServices(ctx, module, func(scoped Runner) error {
			directory := filepath.Join(scoped.Root, module.Directory)
			return announce(output, module.Directory, "coverage", func() error {
				return scoped.runCoverage(ctx, output, directory, module)
			})
		}); err != nil {
			return err
		}
	}
	return nil
}

// Docs validates Markdown navigation and any additional typed documentation
// operation for selected modules.
func (runner Runner) Docs(ctx context.Context, selection []string) error {
	modules, err := runner.selectModules(selection)
	if err != nil {
		return err
	}
	output := runner.Output
	if output == nil {
		output = io.Discard
	}
	for _, module := range modules {
		if !module.Gates["documentation"] {
			_, _ = fmt.Fprintf(output, "[%s] docs: not applicable\n", module.Directory)
			continue
		}
		directory := filepath.Join(runner.Root, module.Directory)
		if err := announce(output, module.Directory, "docs", func() error {
			return runner.checkDocumentation(ctx, directory, module)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) selectModules(selection []string) ([]inventory.Module, error) {
	if len(runner.Catalog.Modules) > inventory.MaximumModules {
		return nil, errors.New("module cardinality limit exceeded")
	}
	available := make(map[string]inventory.Module, len(runner.Catalog.Modules))
	for _, module := range runner.Catalog.Modules {
		available[module.Directory] = module
	}
	unique := make(map[string]struct{}, len(selection))
	for _, directory := range selection {
		if _, ok := available[directory]; !ok {
			return nil, fmt.Errorf("unknown module: %s", directory)
		}
		unique[directory] = struct{}{}
	}
	directories := make([]string, 0, len(unique))
	for directory := range unique {
		directories = append(directories, directory)
	}
	sort.Strings(directories)
	modules := make([]inventory.Module, 0, len(directories))
	for _, directory := range directories {
		modules = append(modules, available[directory])
	}
	return modules, nil
}

func (runner Runner) checkModule(ctx context.Context, output io.Writer, module inventory.Module) error {
	directory := filepath.Join(runner.Root, module.Directory)
	if err := announce(output, module.Directory, "format-check", func() error {
		return runner.checkFormatting(ctx, directory)
	}); err != nil {
		return err
	}
	if err := runner.command(ctx, output, module.Directory, "tidy-check", directory, "mod", "tidy", "-diff"); err != nil {
		return err
	}
	if err := announce(output, module.Directory, "safety", func() error {
		return checkSafety(directory)
	}); err != nil {
		return err
	}
	if module.Gates["lint"] {
		if err := runner.command(ctx, output, module.Directory, "vet", directory, "vet", "./..."); err != nil {
			return err
		}
	}
	if module.Gates["tests"] {
		args := testArguments(module.TestTags, false)
		if err := runner.command(ctx, output, module.Directory, "test", directory, args...); err != nil {
			return err
		}
		if operation, exists := runner.operation(module.Directory, "test"); exists {
			if err := runner.runOperation(ctx, directory, module, operation); err != nil {
				return err
			}
		}
	}
	if module.Gates["race"] {
		args := testArguments(module.TestTags, true)
		if err := runner.command(ctx, output, module.Directory, "race", directory, args...); err != nil {
			return err
		}
	}
	if module.Gates["coverage"] {
		if err := announce(output, module.Directory, "coverage", func() error {
			return runner.runCoverage(ctx, output, directory, module)
		}); err != nil {
			return err
		}
	}
	if module.Gates["mutation"] {
		if err := runner.verifyMutation(ctx, output, module); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		if err := runner.goTool(ctx, output, module.Directory, "lint", directory,
			"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangCILintVersion,
			"run", "--allow-parallel-runners", "--timeout=10m", "./..."); err != nil {
			return err
		}
		if err := runner.goTool(ctx, output, module.Directory, "staticcheck", directory,
			"honnef.co/go/tools/cmd/staticcheck@"+staticcheckVersion, "./..."); err != nil {
			return err
		}
	}
	if operation, exists := runner.operation(module.Directory, "fuzz"); exists {
		if err := announce(output, module.Directory, "fuzz", func() error {
			return runner.runOperation(ctx, directory, module, operation)
		}); err != nil {
			return err
		}
	}
	if module.Gates["documentation"] {
		if err := announce(output, module.Directory, "docs", func() error {
			return runner.checkDocumentation(ctx, directory, module)
		}); err != nil {
			return err
		}
	}
	if module.Gates["api_compatibility"] {
		if operation, exists := runner.operation(module.Directory, "api"); exists {
			if err := announce(output, module.Directory, "api", func() error {
				return runner.runOperation(ctx, directory, module, operation)
			}); err != nil {
				return err
			}
		} else if err := announce(output, module.Directory, "api", func() error {
			return runner.apiModule(ctx, output, module, false)
		}); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		_, _ = fmt.Fprintf(output, "[%s] nilaway\n", module.Directory)
		err := runner.Executor.Run(ctx, Command{
			Name: "go", Dir: directory, Env: map[string]string{"GOWORK": "off"},
			Args: []string{"run", "go.uber.org/nilaway/cmd/nilaway@" + nilAwayVersion,
				"-include-pkgs=" + module.ModulePath, "./..."},
		})
		if err != nil {
			_, _ = fmt.Fprintf(output, "[%s] NilAway advisory: %v\n", module.Directory, err)
		}
	}
	for _, gate := range []string{"conformance", "interoperability", "benchmark"} {
		operation, exists := runner.operation(module.Directory, gate)
		if !exists {
			continue
		}
		if err := announce(output, module.Directory, gate, func() error {
			return runner.runOperation(ctx, directory, module, operation)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) checkModuleLocal(ctx context.Context, output io.Writer, module inventory.Module) error {
	directory := filepath.Join(runner.Root, module.Directory)
	if err := announce(output, module.Directory, "format-check", func() error {
		return runner.checkFormatting(ctx, directory)
	}); err != nil {
		return err
	}
	if err := runner.command(ctx, output, module.Directory, "tidy-check", directory, "mod", "tidy", "-diff"); err != nil {
		return err
	}
	if err := announce(output, module.Directory, "safety", func() error {
		return checkSafety(directory)
	}); err != nil {
		return err
	}
	if module.Gates["lint"] {
		if err := runner.command(ctx, output, module.Directory, "vet", directory, "vet", "./..."); err != nil {
			return err
		}
	}
	if module.Gates["tests"] {
		if err := runner.command(ctx, output, module.Directory, "test", directory, testArguments(module.TestTags, false)...); err != nil {
			return err
		}
	}
	if module.Gates["lint"] {
		if err := runner.goTool(ctx, output, module.Directory, "lint", directory,
			"github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+golangCILintVersion,
			"run", "--allow-parallel-runners", "--timeout=10m", "./..."); err != nil {
			return err
		}
		if err := runner.goTool(ctx, output, module.Directory, "staticcheck", directory,
			"honnef.co/go/tools/cmd/staticcheck@"+staticcheckVersion, "./..."); err != nil {
			return err
		}
	}
	if module.Gates["documentation"] {
		if err := announce(output, module.Directory, "docs-local", func() error {
			return docscheck.CheckWithinContext(ctx, runner.Root, directory)
		}); err != nil {
			return err
		}
	}
	if module.Gates["api_compatibility"] {
		if err := announce(output, module.Directory, "api", func() error {
			return runner.apiModule(ctx, output, module, false)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) createGitleaksConfig() (string, func() error, error) {
	return runner.createOwnedPolicy("gitleaks", "gitleaks-config-*.toml", gitleaksPolicy)
}

func (runner Runner) createAnalysisConfig() (string, func() error, error) {
	return runner.createOwnedPolicy("analysis", "analysis-security-*.yaml", analysisSecurityPolicy)
}

func (runner Runner) createOwnedPolicy(name, pattern, policy string) (string, func() error, error) {
	workspace, ok := runner.Executor.(taskWorkspace)
	if !ok || !filepath.IsAbs(workspace.TemporaryDirectory()) {
		return "", nil, errors.New("security policy requires an absolute task-owned temporary directory")
	}
	files := runner.secretConfigFiles
	if files == nil {
		files = operatingSecretConfigFiles{}
	}
	temporary, err := files.CreateTemp(workspace.TemporaryDirectory(), pattern)
	if err != nil {
		return "", nil, fmt.Errorf("create temporary %s config: %w", name, err)
	}
	path := temporary.Name()
	cleanup := func() error { return files.Remove(path) }
	if _, err := io.WriteString(temporary, policy); err != nil {
		return "", nil, errors.Join(fmt.Errorf("write temporary %s config: %w", name, err), temporary.Close(), cleanup())
	}
	if err := temporary.Close(); err != nil {
		return "", nil, errors.Join(fmt.Errorf("close temporary %s config: %w", name, err), cleanup())
	}
	return path, cleanup, nil
}

func (runner Runner) checkSecurity(ctx context.Context, output io.Writer, directory string, module inventory.Module) error {
	if err := runner.checkSecurityPolicy(ctx, output, directory, module.Directory); err != nil {
		return err
	}
	return runner.runSecurity(ctx, output, directory, module)
}

func (runner Runner) checkSecurityPolicy(ctx context.Context, output io.Writer, directory, module string) error {
	return announce(output, module, "security-suppressions", func() error {
		return checkSecuritySuppressionsBounded(ctx, directory, runner.sourceLimits().entries)
	})
}

func (runner Runner) preflightSecurityPolicies(ctx context.Context, output io.Writer, modules []inventory.Module) error {
	securityEnabled := slices.ContainsFunc(modules, func(module inventory.Module) bool {
		return module.Gates["security"]
	})
	if securityEnabled {
		if err := rejectRepositoryGitleaksIgnore(runner.Root); err != nil {
			return err
		}
	}
	for _, module := range modules {
		if !module.Gates["security"] {
			continue
		}
		directory := filepath.Join(runner.Root, module.Directory)
		if err := runner.checkSecurityPolicy(ctx, output, directory, module.Directory); err != nil {
			return err
		}
	}
	return nil
}

func (runner Runner) runSecurity(ctx context.Context, output io.Writer, directory string, module inventory.Module) error {
	if err := runner.securityTool(ctx, output, module.Directory, "vulnerability", directory,
		"golang.org/x/vuln/cmd/govulncheck@"+govulncheckVersion, "./..."); err != nil {
		return err
	}
	packages, err := runner.gosecPackages(ctx, directory)
	if err != nil {
		return err
	}
	gosecArguments := append([]string{"-nosec-require-rules", "-nosec-require-justification"}, packages...)
	if err := runner.securityTool(ctx, output, module.Directory, "gosec", directory,
		"github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion,
		gosecArguments...); err != nil {
		return err
	}
	analysisPath, cleanupAnalysis, err := runner.createAnalysisConfig()
	if err != nil {
		return err
	}
	if err := runner.securityTool(ctx, output, module.Directory, "owned-security-analysis", directory,
		"github.com/faustbrian/go-analysis/cmd/golib-analysis@"+goAnalysisVersion,
		"check", "-config", analysisPath, "-root", directory, "./..."); err != nil {
		return errors.Join(err, cleanupAnalysis())
	}
	if err := cleanupAnalysis(); err != nil {
		return fmt.Errorf("remove temporary analysis config: %w", err)
	}
	if !runner.repositorySecretsScanned {
		if err := runner.runRepositorySecrets(ctx, output, module.Directory); err != nil {
			return err
		}
	}
	licenseOwner := module.ModulePath
	if repository := strings.TrimSuffix(runner.Catalog.Repository, "/"); repository != "" &&
		(module.ModulePath == repository || strings.HasPrefix(module.ModulePath, repository+"/")) {
		licenseOwner = repository
	}
	if err := runner.securityTool(ctx, output, module.Directory, "licenses", directory,
		"github.com/google/go-licenses/v2@"+goLicensesVersion,
		"check", "./...", "--ignore", licenseOwner); err != nil {
		return err
	}
	return announce(output, module.Directory, "SBOM", func() error { return runner.runSBOM(ctx, directory) })
}

func (runner *Runner) scanSelectedSecurity(ctx context.Context, output io.Writer, modules []inventory.Module) error {
	// Finish every selected source/graph scan before any repository-controlled
	// test can mutate inputs for another module. Runtime service scopes remain
	// owned by the later per-module execution pass.
	for _, module := range modules {
		if module.Gates["security"] {
			if err := runner.runRepositorySecrets(ctx, output, module.Directory); err != nil {
				return err
			}
			runner.repositorySecretsScanned = true
			break
		}
	}
	for _, module := range modules {
		if module.Gates["security"] {
			if err := runner.runSecurity(ctx, output, filepath.Join(runner.Root, module.Directory), module); err != nil {
				return err
			}
		}
	}
	return nil
}

func (runner Runner) runRepositorySecrets(ctx context.Context, output io.Writer, attribution string) error {
	configPath, cleanupSecrets, err := runner.createGitleaksConfig()
	if err != nil {
		return err
	}
	if err := rejectRepositoryGitleaksIgnore(runner.Root); err != nil {
		return errors.Join(err, cleanupSecrets())
	}
	sources, cleanupSources, err := runner.createGitleaksSources(ctx)
	if err != nil {
		return errors.Join(err, cleanupSecrets())
	}
	if err := runner.securityTool(ctx, output, attribution, "secrets-history", sources.history,
		"github.com/zricethezav/gitleaks/v8@"+gitleaksVersion,
		"git", ".", "--config", configPath, "--log-opts=--all", "--ignore-gitleaks-allow",
		"--gitleaks-ignore-path", sources.ignoreRoot, "--no-banner", "--redact"); err != nil {
		return errors.Join(err, cleanupSources(), cleanupSecrets())
	}
	if err := runner.securityTool(ctx, output, attribution, "secrets-current-tree", sources.current,
		"github.com/zricethezav/gitleaks/v8@"+gitleaksVersion,
		"dir", ".", "--config", configPath, "--ignore-gitleaks-allow",
		"--gitleaks-ignore-path", sources.ignoreRoot, "--no-banner", "--redact"); err != nil {
		return errors.Join(err, cleanupSources(), cleanupSecrets())
	}
	if err := cleanupSources(); err != nil {
		return errors.Join(fmt.Errorf("remove temporary gitleaks sources: %w", err), cleanupSecrets())
	}
	if err := cleanupSecrets(); err != nil {
		return fmt.Errorf("remove temporary gitleaks config: %w", err)
	}
	return nil
}

func rejectRepositoryGitleaksIgnore(root string) error {
	info, err := os.Lstat(filepath.Join(root, ".gitleaksignore"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return errors.New("inspect repository-owned .gitleaksignore")
	}
	if info.IsDir() {
		return nil
	}
	return errRepositoryGitleaksIgnore
}

func checkSecuritySuppressions(root string) error {
	return checkSecuritySuppressionsBounded(context.Background(), root, maximumSecuritySourceFiles)
}

func checkSecuritySuppressionsBounded(ctx context.Context, root string, entryLimit int) error {
	sourceRoot, err := os.OpenRoot(root)
	if err != nil {
		return err
	}
	defer sourceRoot.Close()
	return walkSecuritySource(ctx, root, entryLimit, func(relative string, entry os.DirEntry) error {
		filePath := filepath.Join(root, relative)
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == ".golib-tooling" || entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != ".go" {
			return nil
		}
		content, err := readSecuritySuppressionSource(sourceRoot, relative, entry, maximumSecuritySourceSize, securitySuppressionSourceFiles{
			info:  os.DirEntry.Info,
			open:  (*os.Root).Open,
			close: (*os.File).Close,
		})
		if err != nil {
			return err
		}
		fileSet := token.NewFileSet()
		parsed, err := parser.ParseFile(fileSet, filePath, content, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, group := range parsed.Comments {
			for _, suppression := range nativeNosecGroupDirectives(group, fileSet) {
				directive, lineNumber := suppression.arguments, suppression.line
				if err := validateNativeSecurityDirective(directive); err != nil {
					return fmt.Errorf("%s:%d: %w", filepath.ToSlash(filePath), lineNumber, err)
				}
			}
			for _, comment := range group.List {
				lineNumber := fileSet.Position(comment.Slash).Line
				if directive, found := gosecDisableDirective(comment.Text); found {
					if err := validateNativeSecurityDirective(directive); err != nil {
						return fmt.Errorf("%s:%d: %w", filepath.ToSlash(filePath), lineNumber, err)
					}
				}
				if reason, found := nolintGosecDirective(comment.Text); found && strings.TrimSpace(reason) == "" {
					return fmt.Errorf("%s:%d: nolint:gosec requires an inline reason", filepath.ToSlash(filePath), lineNumber)
				}
			}
		}
		return nil
	})
}

// Collaborators belong to one source read and retain real confined roots and
// file handles. The production caller supplies the fixed source-size limit.
type securitySuppressionSourceFiles struct {
	info  func(os.DirEntry) (os.FileInfo, error)
	open  func(*os.Root, string) (*os.File, error)
	close func(*os.File) error
}

func readSecuritySuppressionSource(root *os.Root, relative string, entry os.DirEntry, limit int64, files securitySuppressionSourceFiles) ([]byte, error) {
	info, err := files.info(entry)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("security suppression source is not a regular file: %s", entry.Name())
	}
	if info.Size() > limit {
		return nil, fmt.Errorf("security suppression source exceeds size limit: %s", entry.Name())
	}
	file, err := files.open(root, relative)
	if err != nil {
		return nil, err
	}
	content, readErr := io.ReadAll(io.LimitReader(file, limit+1))
	closeErr := files.close(file)
	if readErr != nil || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}
	if int64(len(content)) > limit {
		return nil, fmt.Errorf("security suppression source exceeds size limit: %s", entry.Name())
	}
	return content, nil
}

func nativeSecurityDirective(line string, allowGosecDisable bool) (string, bool) {
	body := strings.TrimSpace(line)
	if allowGosecDisable {
		if directive, found := gosecDisableDirective(body); found {
			return directive, true
		}
	}
	if trimmed, found := strings.CutPrefix(body, "//"); found {
		body = strings.TrimSpace(trimmed)
	}
	const nosec = "#nosec"
	if trimmed, found := strings.CutPrefix(body, nosec); found {
		return strings.TrimSpace(trimmed), true
	}
	return "", false
}

type nativeSecuritySuppression struct {
	arguments string
	line      int
}

func nativeNosecGroupDirectives(group *ast.CommentGroup, fileSet *token.FileSet) []nativeSecuritySuppression {
	var suppressions []nativeSecuritySuppression
	for _, comment := range group.List {
		body := comment.Text
		baseLine := fileSet.Position(comment.Slash).Line
		if strings.HasPrefix(body, "//") {
			if directive, found := nativeSecurityDirective(body, false); found {
				suppressions = append(suppressions, nativeSecuritySuppression{arguments: directive, line: baseLine})
			}
			continue
		}
		// The caller supplies comments from parser.ParseFile: after line
		// comments above, every remaining scanner token is a block comment.
		body = strings.TrimPrefix(body, "/*")
		body = strings.TrimSuffix(body, "*/")
		for offset, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "//") {
				continue
			}
			if directive, found := nativeSecurityDirective(line, false); found {
				suppressions = append(suppressions, nativeSecuritySuppression{arguments: directive, line: baseLine + offset})
			}
		}
	}
	return suppressions
}

func gosecDisableDirective(comment string) (string, bool) {
	body := strings.TrimSpace(comment)
	const directive = "//gosec:disable"
	if body == directive || strings.HasPrefix(body, directive+" ") {
		return strings.TrimSpace(strings.TrimPrefix(body, directive)), true
	}
	return "", false
}

func nolintGosecDirective(comment string) (string, bool) {
	body := strings.TrimLeft(comment, "/ ")
	if !strings.HasPrefix(body, "nolint:") {
		return "", false
	}
	directive, reason, _ := strings.Cut(body, "//")
	for linter := range strings.SplitSeq(strings.TrimPrefix(directive, "nolint:"), ",") {
		if strings.EqualFold(strings.TrimSpace(linter), "gosec") {
			return reason, true
		}
	}
	return "", false
}

func validateNativeSecurityDirective(arguments string) error {
	rules, reason, found := strings.Cut(arguments, "--")
	rules = strings.TrimSpace(rules)
	if !securityRuleList.MatchString(rules) {
		fields := strings.Fields(arguments)
		if !found && len(fields) > 0 && securityRuleList.MatchString(fields[0]) {
			return errors.New("security suppression requires a reason after --")
		}
		return errors.New("security suppression requires exact rule IDs")
	}
	if !found || strings.TrimSpace(strings.TrimLeft(reason, "-")) == "" {
		return errors.New("security suppression requires a reason after --")
	}
	return nil
}

func (runner Runner) goTool(ctx context.Context, output io.Writer, module, gate, directory, tool string, args ...string) error {
	if tool == "honnef.co/go/tools/cmd/staticcheck@"+staticcheckVersion {
		return runner.staticcheck(ctx, output, module, gate, directory, args...)
	}
	arguments := append([]string{"run", tool}, args...)
	return runner.command(ctx, output, module, gate, directory, arguments...)
}

func (runner Runner) securityTool(ctx context.Context, output io.Writer, module, gate, directory, tool string, args ...string) (result error) {
	gitleaks := tool == "github.com/zricethezav/gitleaks/v8@"+gitleaksVersion
	gosec := tool == "github.com/securego/gosec/v2/cmd/gosec@"+gosecVersion
	if gosec {
		args = append([]string{"-fmt=json"}, args...)
	}
	if gitleaks {
		path, cleanup, err := runner.createOwnedPolicy("gitleaks metadata", "gitleaks-metadata-*.tmpl", secretMetadataTemplate)
		if err != nil {
			return err
		}
		defer func() { result = errors.Join(result, cleanup()) }()
		args = append(slices.Clone(args), "--exit-code=42", "--report-format=template", "--report-path=-", "--report-template", path)
	}
	arguments := append([]string{"run", tool}, args...)
	return announce(output, module, gate, func() error {
		scan := func(binary string) error {
			stdout := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
			stderr := &boundedProcessOutput{limit: maximumSecurityProcessOutput}
			var stdoutWriter io.Writer = stdout
			var report *sourceInventoryOutput
			if gosec {
				report = &sourceInventoryOutput{}
				report.limit = maximumSecurityProcessOutput
				stdout = &report.boundedProcessOutput
				stdoutWriter = report
				defer func() { clear(report.data.Bytes()); report.data.Reset() }()
			}
			if gitleaks {
				stdout.metadata = &secretMetadata{}
				// Go run reports the child status as its final stderr line while itself
				// returning status 1. Gitleaks reserves 42 here for completed findings;
				// its setup and scanning errors still use status 1.
				stderr.terminalLine = "exit status 42"
			}
			name, scannerArguments := "go", arguments
			if binary != "" {
				name, scannerArguments = binary, args
			}
			err := runner.Executor.Run(ctx, Command{
				Name: name, Args: scannerArguments, Dir: directory, Env: map[string]string{"GOWORK": "off"}, boundedScanner: true,
				Stdout: stdoutWriter, Stderr: stderr,
			})
			var overflow error
			if stdout.didOverflow() || stderr.didOverflow() {
				overflow = fmt.Errorf("security scanner output exceeded %d bytes", maximumSecurityProcessOutput)
			}
			if err != nil || overflow != nil {
				var classification error
				if err != nil && overflow == nil && ctx.Err() == nil && gitleaks && stderr.matchedTerminalLine() {
					classification = errors.New("secret-findings" + stdout.findingMetadata())
				}
				if err != nil && overflow == nil && ctx.Err() == nil && gosec {
					metadata := "gosec-tool-or-report-failure"
					var status interface{ ExitCode() int }
					if errors.As(err, &status) && status.ExitCode() == 1 {
						metadata = gosecFailureMetadata(report.data.Bytes(), directory)
					}
					classification = errors.New(metadata)
				}
				return fmt.Errorf("%s %s: %w", module, gate, errors.Join(overflow, err, classification))
			}
			return nil
		}
		if tool, ok := securityCompilerTool(tool); ok {
			if err := runner.withCompilerTool(ctx, tool, scan); err != nil {
				return fmt.Errorf("%s %s: %w", module, gate, err)
			}
			return nil
		}
		return scan("")
	})
}

func (runner Runner) runCoverage(ctx context.Context, output io.Writer, directory string, module inventory.Module) error {
	targets := make([]string, 0, len(module.Packages))
	for _, packagePolicy := range module.Packages {
		if packagePolicy.CoverageRequired {
			targets = append(targets, packagePolicy.ImportPath)
		}
	}
	slices.Sort(targets)
	if len(targets) == 0 {
		return errors.New("coverage gate has no coverage-required packages")
	}
	files := runner.coverageFiles
	if files == nil {
		files = operatingCoverageFiles{}
	}
	temporaryDirectory := ""
	if workspace, ok := runner.Executor.(taskWorkspace); ok {
		temporaryDirectory = workspace.TemporaryDirectory()
	}
	profile, err := files.CreateTemp(temporaryDirectory)
	if err != nil {
		return fmt.Errorf("create coverage profile: %w", err)
	}
	profilePath := profile.Name()
	if err := profile.Close(); err != nil {
		_ = files.Remove(profilePath)
		return fmt.Errorf("close coverage profile: %w", err)
	}
	defer func() { _ = files.Remove(profilePath) }()
	args := []string{"test"}
	if len(module.TestTags) > 0 {
		args = append(args, "-tags="+strings.Join(module.TestTags, ","))
	}
	args = append(args, "./...", "-count=1", "-timeout=20m", "-covermode=atomic", "-coverpkg="+strings.Join(targets, ","), "-coverprofile="+profilePath)
	if err := runner.Executor.Run(ctx, Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
		return err
	}
	opened, err := files.Open(profilePath)
	if err != nil {
		return fmt.Errorf("open coverage profile: %w", err)
	}
	defer opened.Close()
	mode := "exact"
	for _, acceptance := range runner.Policy.Coverage.Modules {
		if acceptance.Module == module.Directory {
			mode = acceptance.Mode
			break
		}
	}
	report, err := coverage.VerifyWithMode(opened, targets, mode)
	_, _ = io.WriteString(output, report)
	if err != nil {
		return err
	}
	if mode == "evidence" {
		_, _ = io.WriteString(output, "coverage evidence collected; adequacy and release readiness are not certified\n")
	} else {
		_, _ = io.WriteString(output, "all production packages have exact 100% statement coverage\n")
	}
	return nil
}

func (runner Runner) operation(module, gate string) (config.Operation, bool) {
	for _, operation := range runner.Policy.Operations {
		if operation.Module == module && operation.Gate == gate {
			return operation, true
		}
	}
	return config.Operation{}, false
}

func (runner Runner) runOperation(ctx context.Context, directory string, module inventory.Module, operation config.Operation) error {
	for index, step := range operation.Steps {
		timeout, err := time.ParseDuration(step.Timeout)
		if err != nil {
			return fmt.Errorf("%s %s step %d timeout: %w", module.Directory, operation.Gate, index, err)
		}
		stepContext, cancel := context.WithTimeout(ctx, timeout)
		command, err := operationCommand(directory, module, step)
		if err == nil {
			err = runner.Executor.Run(stepContext, command)
		}
		cancel()
		if err != nil {
			return fmt.Errorf("%s %s step %d: %w", module.Directory, operation.Gate, index, err)
		}
	}
	return nil
}

func operationCommand(directory string, module inventory.Module, step config.Step) (Command, error) {
	switch step.Type {
	case "go-test":
		args := []string{"test"}
		if len(module.TestTags) > 0 {
			args = append(args, "-tags="+strings.Join(module.TestTags, ","))
		}
		args = append(args, step.Packages...)
		args = append(args, fmt.Sprintf("-count=%d", step.Count), "-timeout="+step.Timeout)
		if step.Run != "" {
			args = append(args, "-run="+step.Run)
		}
		if step.Benchmark != "" {
			args = append(args, "-run=^$", "-bench="+step.Benchmark, "-benchmem", "-benchtime="+step.Budget)
		}
		if step.Fuzz != "" {
			args = append(args, "-run=^$", "-fuzz="+step.Fuzz, "-fuzztime="+step.Budget)
		}
		return Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}, nil
	case "make":
		makefile, err := repositoryfile.Read(directory, step.Makefile, maximumMakefileSize)
		if err != nil {
			return Command{}, fmt.Errorf("read makefile: %w", err)
		}
		return Command{
			Name:  "make",
			Args:  []string{"--no-print-directory", "-f", "-", step.Target},
			Dir:   directory,
			Env:   map[string]string{"GOWORK": "off"},
			Stdin: bytes.NewReader(makefile),
		}, nil
	default:
		return Command{}, fmt.Errorf("unsupported operation type: %s", step.Type)
	}
}

func testArguments(tags []string, race bool) []string {
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	return append(args, "./...", "-count=1", "-timeout=20m")
}

func (runner Runner) command(ctx context.Context, output io.Writer, module, gate, directory string, args ...string) error {
	return announce(output, module, gate, func() error {
		if err := runner.Executor.Run(ctx, Command{Name: "go", Args: args, Dir: directory, Env: map[string]string{"GOWORK": "off"}}); err != nil {
			return fmt.Errorf("%s %s: %w", module, gate, err)
		}
		return nil
	})
}

func announce(output io.Writer, module, gate string, operation func() error) error {
	_, _ = fmt.Fprintf(output, "[%s] %s\n", module, gate)
	return operation()
}

func (runner Runner) checkFormatting(ctx context.Context, root string) error {
	files := make([]string, 0)
	if err := walkModuleFiles(root, func(path string, _ fs.DirEntry) error {
		relative := filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator)))
		if strings.ContainsAny(relative, "\r\n") {
			return fmt.Errorf("go source path contains a line break: %q", relative)
		}
		files = append(files, relative)
		return nil
	}); err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	sort.Strings(files)
	var output boundedBuffer
	if err := runner.Executor.Run(ctx, Command{
		Name: "gofmt", Args: append([]string{"-l", "--"}, files...), Dir: root,
		Stdout: &output,
	}); err != nil {
		return fmt.Errorf("run gofmt: %w", err)
	}
	if output.overflow {
		return errors.New("gofmt output exceeds limit")
	}
	unformatted := strings.TrimSpace(output.String())
	if unformatted == "" {
		return nil
	}
	return fmt.Errorf("unformatted Go file: %s", strings.Split(unformatted, "\n")[0])
}

func checkSafety(root string) error {
	return walkModuleFiles(root, func(path string, _ fs.DirEntry) error {
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly|parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, imported := range parsed.Imports {
			if imported.Path.Value == `"unsafe"` || imported.Path.Value == `"C"` {
				return fmt.Errorf("forbidden production import %s in %s", imported.Path.Value, path)
			}
		}
		for _, group := range parsed.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:linkname") {
					return fmt.Errorf("forbidden go:linkname directive in %s", path)
				}
			}
		}
		return nil
	})
}

func walkModuleFiles(root string, visit func(string, fs.DirEntry) error) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root {
				name := entry.Name()
				if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
					return filepath.SkipDir
				}
				if _, statErr := os.Stat(filepath.Join(path, "go.mod")); statErr == nil {
					return filepath.SkipDir
				} else if !os.IsNotExist(statErr) {
					return statErr
				}
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		return visit(path, entry)
	})
}
