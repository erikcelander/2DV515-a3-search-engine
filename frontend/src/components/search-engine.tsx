'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  TableHead,
  TableRow,
  TableHeader,
  TableCell,
  TableBody,
  Table,
} from '@/components/ui/table'
import Link from 'next/link'

type SearchResult = {
  url: string
  contentScore: number
  locationScore: number
  pageRankScore: number
  totalScore: number
}

export function SearchEngine() {
  const [searchQuery, setSearchQuery] = useState('')
  const [results, setResults] = useState<SearchResult[]>([])
  const [elapsedTime, setElapsedTime] = useState(0)

  const handleSearch = async () => {
    const startTime = performance.now()
    try {
      const response = await fetch('http://localhost:8080/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ word: searchQuery }),
      })

      if (!response.ok) {
        console.error('Error:', response.statusText)
        setResults([])
      }

      const data: SearchResult[] = await response.json()
      setResults(data || [])
    } catch (error) {
      console.error('Error:', error)
      setResults([])
    } 
    const endTime = performance.now()
    setElapsedTime((endTime - startTime) / 1000)
  }

  return (
    <div className='min-h-screen bg-gray-900 text-white p-8'>
      <div className='flex flex-col gap-6'>
        <div className='flex gap-4 items-center'>
          <label className='text-lg' htmlFor='search'>
            Search articles:
          </label>
          <input
            className='flex-1 px-4 py-2 bg-gray-800 border border-gray-700 rounded-md focus:outline-none focus:border-blue-500'
            id='search'
            placeholder='java programming'
            type='text'
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <Button className='bg-blue-600 hover:bg-blue-700 text-white' onClick={handleSearch}>
            Search
          </Button>
        </div>
        {results.length > 0 ? (
          <>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className='text-left bg-gray-800 w-1/2'>Link</TableHead>
                  <TableHead className='text-center bg-gray-800 w-1/8'>Content</TableHead>
                  <TableHead className='text-center bg-gray-800 w-1/8'>Location</TableHead>
                  <TableHead className='text-center bg-gray-800 w-1/8'>PageRank</TableHead>
                  <TableHead className='text-center font-bold bg-gray-800 w-1/8'>Score</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {results.slice(0, 5).map((result, index) => (
                  <TableRow key={index}>
                    <TableCell className='text-blue-500 text-left font-medium w-1/2'>
                      <Link href={`https://wikipedia.org/wiki/${result.url}`}> 
                      {result.url}
                      </Link>
                    </TableCell>
                    <TableCell className='text-center w-1/8'>
                      {result.contentScore.toFixed(2)}
                    </TableCell>
                    <TableCell className='text-center w-1/8'>
                      {result.locationScore.toFixed(2)}
                    </TableCell>
                    <TableCell className='text-center w-1/8'>
                      {result.pageRankScore.toFixed(2)}
                    </TableCell>
                    <TableCell className='font-bold text-center w-1/8'>
                      {result.totalScore.toFixed(2)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            <div className='text-sm'>
              Found {results.length} results in {elapsedTime.toFixed(3)}s
            </div>
          </>
        ) : (
          <></>
        )}
      </div>
    </div>
  )
}
