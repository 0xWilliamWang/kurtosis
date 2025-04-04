package directory

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"math"
)

const (
	DirectoryTypeName = "Directory"

	ArtifactNamesAttr = "artifact_names"
	PersistentKeyAttr = "persistent_key"
	SizeKeyAttr       = "size"

	atleastOneMegabyte       = 1
	megaByteToByteMultiplier = 1024 * 1024
)

func NewDirectoryType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: DirectoryTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ArtifactNamesAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.List],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.StringListWithNotEmptyValues(value, ArtifactNamesAttr)
					},
				},
				{
					Name:              PersistentKeyAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ArtifactNamesAttr)
					},
				},
				{
					Name:              SizeKeyAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Int64InRange(value, SizeKeyAttr, atleastOneMegabyte, math.MaxInt64)
					},
				},
			},
		},

		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(DirectoryTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &Directory{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type Directory struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

// CreateDirectoryFromFilesArtifact creates a directory object from a file artifact name. This is only for backward
// compatibility
func CreateDirectoryFromFilesArtifact(
	filesArtifactName string,
) (*Directory, *startosis_errors.InterpretationError) {
	args := []starlark.Value{
		starlark.String(filesArtifactName),
		nil,
		nil,
	}

	argumentDefinitions := NewDirectoryType().Arguments
	argumentValuesSet := builtin_argument.NewArgumentValuesSet(argumentDefinitions, args)
	kurtosisDefaultValue, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ArtifactNamesAttr, argumentValuesSet)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &Directory{
		KurtosisValueTypeDefault: kurtosisDefaultValue,
	}, nil
}

func (directory *Directory) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := directory.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &Directory{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (directory *Directory) GetArtifactNamesIfSet() ([]string, bool, *startosis_errors.InterpretationError) {
	artifactNames := []string{}
	fileArtifacts, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.List](
		directory.KurtosisValueTypeDefault, ArtifactNamesAttr)
	if !found {
		return nil, false, nil
	}
	if interpretationErr != nil {
		var fileArtifact starlark.String
		fileArtifact, found, interpretationErr = kurtosis_type_constructor.ExtractAttrValue[starlark.String](
			directory.KurtosisValueTypeDefault, ArtifactNamesAttr)
		if interpretationErr != nil {
			return nil, false, interpretationErr
		}
		if !found {
			return nil, false, nil
		}
		artifactNames = append(artifactNames, fileArtifact.GoString())
	} else {
		if fileArtifacts.Len() > 0 {
			artifactNames, interpretationErr = kurtosis_types.SafeCastToStringSlice(fileArtifacts, ArtifactNamesAttr)
			if interpretationErr != nil {
				return nil, false, interpretationErr
			}
		}
	}

	return artifactNames, true, nil
}

func (directory *Directory) GetPersistentKeyIfSet() (string, bool, *startosis_errors.InterpretationError) {
	persistentKey, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		directory.KurtosisValueTypeDefault, PersistentKeyAttr)
	if interpretationErr != nil {
		return "", false, interpretationErr
	}
	if !found {
		return "", false, nil
	}
	return persistentKey.GoString(), true, nil
}

func (directory *Directory) GetSizeOrDefault() (int64, *startosis_errors.InterpretationError) {
	size, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		directory.KurtosisValueTypeDefault, SizeKeyAttr)
	if interpretationErr != nil {
		return 0, interpretationErr
	}
	if !found {
		return startosis_constants.DefaultPersistentDirectorySize, nil
	}
	sizeInt64, ok := size.Int64()
	if !ok {
		return 0, startosis_errors.NewInterpretationError("Couldn't convert size '%v' to int64", size)
	}
	return sizeInt64 * megaByteToByteMultiplier, nil
}
