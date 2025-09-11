/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleProtobufDuration from "../protobuf/duration.pb"
import * as GoogleApiLaunch_stage from "./launch_stage.pb"

export enum ClientLibraryOrganization {
  CLIENT_LIBRARY_ORGANIZATION_UNSPECIFIED = "CLIENT_LIBRARY_ORGANIZATION_UNSPECIFIED",
  CLOUD = "CLOUD",
  ADS = "ADS",
  PHOTOS = "PHOTOS",
  STREET_VIEW = "STREET_VIEW",
  SHOPPING = "SHOPPING",
  GEO = "GEO",
  GENERATIVE_AI = "GENERATIVE_AI",
}

export enum ClientLibraryDestination {
  CLIENT_LIBRARY_DESTINATION_UNSPECIFIED = "CLIENT_LIBRARY_DESTINATION_UNSPECIFIED",
  GITHUB = "GITHUB",
  PACKAGE_MANAGER = "PACKAGE_MANAGER",
}

export type CommonLanguageSettings = {
  reference_docs_uri?: string
  destinations?: ClientLibraryDestination[]
  selective_gapic_generation?: SelectiveGapicGeneration
}

export type ClientLibrarySettings = {
  version?: string
  launch_stage?: GoogleApiLaunch_stage.LaunchStage
  rest_numeric_enums?: boolean
  java_settings?: JavaSettings
  cpp_settings?: CppSettings
  php_settings?: PhpSettings
  python_settings?: PythonSettings
  node_settings?: NodeSettings
  dotnet_settings?: DotnetSettings
  ruby_settings?: RubySettings
  go_settings?: GoSettings
}

export type Publishing = {
  method_settings?: MethodSettings[]
  new_issue_uri?: string
  documentation_uri?: string
  api_short_name?: string
  github_label?: string
  codeowner_github_teams?: string[]
  doc_tag_prefix?: string
  organization?: ClientLibraryOrganization
  library_settings?: ClientLibrarySettings[]
  proto_reference_documentation_uri?: string
  rest_reference_documentation_uri?: string
}

export type JavaSettings = {
  library_package?: string
  service_class_names?: {[key: string]: string}
  common?: CommonLanguageSettings
}

export type CppSettings = {
  common?: CommonLanguageSettings
}

export type PhpSettings = {
  common?: CommonLanguageSettings
}

export type PythonSettingsExperimentalFeatures = {
  rest_async_io_enabled?: boolean
  protobuf_pythonic_types_enabled?: boolean
  unversioned_package_disabled?: boolean
}

export type PythonSettings = {
  common?: CommonLanguageSettings
  experimental_features?: PythonSettingsExperimentalFeatures
}

export type NodeSettings = {
  common?: CommonLanguageSettings
}

export type DotnetSettings = {
  common?: CommonLanguageSettings
  renamed_services?: {[key: string]: string}
  renamed_resources?: {[key: string]: string}
  ignored_resources?: string[]
  forced_namespace_aliases?: string[]
  handwritten_signatures?: string[]
}

export type RubySettings = {
  common?: CommonLanguageSettings
}

export type GoSettings = {
  common?: CommonLanguageSettings
  renamed_services?: {[key: string]: string}
}

export type MethodSettingsLongRunning = {
  initial_poll_delay?: GoogleProtobufDuration.Duration
  poll_delay_multiplier?: number
  max_poll_delay?: GoogleProtobufDuration.Duration
  total_poll_timeout?: GoogleProtobufDuration.Duration
}

export type MethodSettings = {
  selector?: string
  long_running?: MethodSettingsLongRunning
  auto_populated_fields?: string[]
}

export type SelectiveGapicGeneration = {
  methods?: string[]
  generate_omitted_as_internal?: boolean
}