<?php
namespace App\Tests\Rules;

final class RuleReferenceTest
{
    public function expectedDiagnostic(): string
    {
        return 'App\Rules\OldRule must not be used here';
    }
}
