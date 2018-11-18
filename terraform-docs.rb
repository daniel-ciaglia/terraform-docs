class TerraformDocs < Formula

  desc "Tool to generate documentation from Terraform modules"
  homepage "https://github.com/daniel-ciaglia/terraform-docs"
  url "https://github.com/daniel-ciaglia/terraform-docs/releases/download/0.5.0.1/terraform-docs-darwin-amd64"
  sha256 "9df0ece1038527f4c4d286598b3f848dbc0b667f6f945c944c64e55d20d9fe5d"
  version "0.5.0.1"

  bottle :unneeded

  def install
    bin.install "terraform-docs-darwin-amd64" => "terraform-docs"
  end

  test do
    system "#{bin}/terraform-docs", "--help"
  end

end
