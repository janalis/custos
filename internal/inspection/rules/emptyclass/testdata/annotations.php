<?php
namespace Drupal\demo\Plugin\Field\FieldFormatter {
    /**
     * Plain text formatter.
     *
     * @FieldFormatter(
     *   id = "demo_plain",
     *   label = @Translation("Plain"),
     * )
     */
    class PlainFormatter extends \Base {}

    /**
     * @Annotation
     * @Target("CLASS")
     */
    class Marker {}

    /** @ORM\Entity */
    class Row {}

    /**
     * Only phpDoc tags here.
     *
     * @internal
     * @see Row
     */
    class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Leftover</weak_warning> {}
}
