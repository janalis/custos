<?php
class Registry extends ArrayObject
{
}

function array_as_props(array $data): array
{
    $ao = new ArrayObject($data, ArrayObject::ARRAY_AS_PROPS);
    $it = new \ArrayIterator($data, ArrayIterator::ARRAY_AS_PROPS);
    $reg = new Registry($data, ArrayObject::ARRAY_AS_PROPS);
    return [isset($ao->name), empty($ao->name), isset($it->name), isset($reg->name)];
}
