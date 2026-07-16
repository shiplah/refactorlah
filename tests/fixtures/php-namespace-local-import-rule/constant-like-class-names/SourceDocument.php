<?php
namespace App\Parsing;

final readonly class SourceDocument
{
    public function __construct(
        private XML_READER $reader,
        private __TOKEN__ $token,
    ) {}
}
