desc 'Update dependencies'
task :deps do
  puts '=====> Updating dependencies...'
  sh 'go get -d -v ./...'
  sh "go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs -n1 go get -d"
end

desc 'Build lmk'
task :build => :deps do
  puts '=====> Building...'
  sh 'go build ./...'
end

desc 'Cross compile'
task :cross_compile => :deps do
  raise 'Not supported yet!'
end

desc 'Release'
task :release => :cross_compile do
  raise 'Not supported yet!'
end
