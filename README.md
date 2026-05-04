# VICS

VICS is the initial Go core for a Visual Information Categorization System. It organizes local image assets by exact file hashes, semantic tags, multilingual tag labels, and perceptual hashes for visual similarity search.

The library is intentionally split into small packages:

- `core` defines domain types, store interfaces, query types, and a service wrapper.
- `storage/sqlite` provides the first persistent backend.
- `perceptual` defines hashing interfaces and Hamming-distance utilities.
- `perceptual/goimagehash` adapts `github.com/corona10/goimagehash`.
- `ingest` imports local image files into any `core.Store`.

## Loading Assets And Tags

Create a `storage/sqlite.Store`, wrap it with `core.NewService`, then use `ingest.Importer` to walk a local image directory. The importer computes a SHA-256 exact file hash, reads image dimensions when the format is supported by Go's image decoders, creates a `core.Asset`, and optionally stores perceptual hashes from a configured `perceptual.Hasher`.

Tags are loaded separately through `Store.UpsertTag`, `Store.UpsertTagLabel`, and `Store.AddTagToAsset`. Japanese tag sidecar parsing is left as a future importer extension, because raw tag text should first be mapped to semantic tag concepts.

## Asset, Tag, And TagLabel

An `Asset` is a local image file record. The type is deliberately not named `Image`, so the domain can later support variants, metadata, and non-rendering representations without confusing the model.

A `Tag` is a language-neutral semantic concept, such as `tag_mountain`. A `TagLabel` is human-readable text for that concept in a specific language. For example, one tag can have labels `山` in Japanese, `yama` in Romaji, and `mountain` in English.

## Exact Hashes And Perceptual Hashes

`Asset.Hash` is the lowercase SHA-256 hash of the file bytes. It identifies exact duplicate files.

`AssetPerceptualHash` stores visual hashes such as aHash, dHash, and pHash. These support near-duplicate and visually-similar image search by comparing bit differences with Hamming distance.

## Include And Exclude Tag Search

`AssetQuery.IncludeTags` requires an asset to have every listed tag. `AssetQuery.ExcludeTags` removes assets that have any listed tag. Text search currently matches asset paths and normalized tag labels.

## Similar Image Search

`SimilarityQuery` can start from an existing asset ID or from a raw perceptual hash hex string. The SQLite backend loads candidate hashes for the same algorithm and computes Hamming distance in Go, keeping the implementation portable and easy to test.

## Why SQLite First

SQLite is a good first backend for a local image library: it is embedded, durable, simple to inspect, and does not require a server. The core package only depends on interfaces, so other backends can be added later without changing domain types or service behavior.
