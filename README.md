# wirepb

This repository implements an experimental wire-format canonical string format
for protocol buffer messages.

Specifically, the `wirepb.Canonical` function implements
[Algorithm 2](#algorithm-2) described below.

## Background

The [Protocol Buffer encoding rules][pbenc] do not guarantee a consistent
binary encoding for any given message.  The encoder is allowed to store field
encodings in any order, regardless of their tag values or their order of
declaration within the protobuf schema.

Even within a single process, the encoder is allowed to produce different
output for repeated encodings of the same message.  Although most encoders do
not encode differently by design, messages that use [maps][pbmap], or that are
encoded with different options, often produce different outputs. This can be
true even if there are no [unknown fields][pbunk].

While a correct encoding must decode to an _equivalent_ message, these rules
mean that hashes, checksums, and fingerprints of wire-format protobuf messages
are not guaranteed to be stable over time.

Applications that need to compute reliable hashes, checksums, or fingerprints
of message structures should generally define them in terms of the schematic
structure of the type.

This is relatively easy to do for a single language, but can be tricky when the
same value needs to be computed across multiple languages that share the
protocol buffer as an interchange format.  The in-memory organization of data
types like strings, integers, floating-point values, arrays, and maps differs
-- so that a rule that is easy to implement in one language may be inefficient
or impractical in another.

Some projects work around this by avoiding "problematic" constructs like maps,
enabling the "deterministic" option in encoders that support it, and relying on
the encoder to respect language-specific properties like array and tag order
when laying out messages in memory. These tactics help, but are insufficient.
Moreover, they often do not generalize across languages.

## Canonical Layout

One possible solution to this dilemma is to define a rule for canonical layout
in the wire format. We can take advantage of the rule that the encoder is
allowed to emit, and the decoder is required to accept, arbitrary ordering of
fields within a message.

In summary, the rule for canonical layout uses the fact that a wire encoding is
a concatenated sequence of `key|value` strings, where the `key` comprises a
field tag number and a "wire type" sufficient to describe the number and layout
of the `value` bytes.

To impose a consistent ordering, consider the following algorithm:

### Algorithm 1

1. Parse the message into a sequence of (type, tag, value) tuples.
2. Sort the tuples by tag (numerically), then value (lexicographically).
3. Rearrange the fields into the resulting order.

Algorithm 1 is not quite sufficient, however: A message field may contain
another message encoded as a string. Furthermore, the wire encoding does not
distinguish between the encoding of a message and the encodings of an arbitrary
byte string.

The only way to know for sure whether a field contains a message or an opaque
string is to consult the schema. However, we can work around this by modifying
the algorithm as follows:

### Algorithm 2

1. Parse the input into a sequence of (type, tag, value) tuples.
   If parsing fails, return the input unmodified. Otherwise:
2. For each tuple whose wire type is "string", apply Algorithm 2 to its value.
3. Sort the resulting tuples by tag (numerically), then value (lexicographically).
4. Rearrange the fields into the resulting order.

Algorithm 2 fixes most of the problems with Algorithm 1. There are three main
issues remaining:

- The algorithm cannot distinguish between a field of opaque string type that
  contains the wire encoding of a message, and a field of message type. Compare:

  ```protobuf
  message M1 {
    bytes content = 1;
  }
  message M2 {
    Foo content = 1;
  }
  ```

  where the `content` field of M1 contains the wire encoding of a `Foo`.
  Algorithm 2 would canonicalize both of these messages identically.

  This is semantically safe: Decoding the result as `M1` and then decoding its
  `content` field would produce the same (valid) `Foo` as decoding the result
  as `M2` and observing its `content` field.

  However, encoding an `M1`, running Algorithm 2, and then decoding the result
  as an `M1`, would not result in an equivalent message.

- The algorithm does not unify default-valued fields with unset (omitted)
  fields.  In [proto3][pb3], the only default values are equivalent to omitting
  the field from the message, and most encoders do omit them. However, if an
  encoder does include default-valued fields in its output, Algorithm 2 will
  generate a different string than if the field was omitted.

- The algorithm does not handle "packed" repeated fields of scalar type.
  Packed repeated fields are encoded as a wire-type string of concatenated
  varint values. The values within the string do not have the structure of a
  message, so the algorithm does not "fix" the order of the packed values.

[pbenc]: https://developers.google.com/protocol-buffers/docs/encoding
[pbfo]: https://developers.google.com/protocol-buffers/docs/encoding#order
[pbmap]: https://developers.google.com/protocol-buffers/docs/proto3#maps
[pbunk]: https://developers.google.com/protocol-buffers/docs/proto3#unknowns
[pb3]: https://developers.google.com/protocol-buffers/docs/proto3
